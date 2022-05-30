package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
	"time"
)

func CheckCommentsOfVideo(videoID uint64) error {
	// 确保 KeycommentsOfVideo存在
	videoIDStr := strconv.FormatUint(videoID, 10)
	keyCommentsOfVideo := "CommentsOfVideo:" + videoIDStr
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyCommentsOfVideo).Result()
	if err != nil {
		return err
	} else if n <= 0 {
		// KeyCommentsOfVideo 不存在
		var commentList []dao.Comment
		result := global.GVAR_DB.Where("video_id = ?", videoID).Find(&commentList)
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("DeleteComment fail")
		}
		// 翻转 commentList: 最近的评论放在前面
		numComments := int(result.RowsAffected)
		for i, j := 0, numComments-1; i < j; i, j = i+1, j-1 {
			commentList[i], commentList[j] = commentList[j], commentList[i]
		}
		err = GoCommentsOfVideo(commentList, keyCommentsOfVideo)
	}
	return err
}

func CheckVideo(videoID uint64) error {
	videoIDStr := strconv.FormatUint(videoID, 10)
	keyVideo := "Video:" + videoIDStr
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyVideo).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// "video_id"不存在
		var video dao.Video
		result := global.GVAR_DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("Redis出错或videoID不存在")
		}
		// 写redis Video:videoID
		err = GoVideo(video)
	}
	return err
}

func GoFeed(allVideos []dao.Video) error {
	var listZ = make([]*redis.Z, 0, len(allVideos))
	for _, video := range allVideos {
		listZ = append(listZ, &redis.Z{float64(video.CreatedAt.UnixMilli()) / 1000, video.VideoID})
	}
	return global.GVAR_REDIS.ZAdd(global.GVAR_CONTEXT, "feed", listZ...).Err()
}

func GoVideo(video dao.Video) error {
	keyVideo := "Video:" + strconv.FormatUint(video.VideoID, 10)
	err := global.GVAR_REDIS.HSet(global.GVAR_CONTEXT, keyVideo, "title", video.Title, "play_name", video.PlayName, "cover_name", video.CoverName,
		"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount, "author_id", video.AuthorID, "created_at", video.CreatedAt.UnixMilli()).Err()
	return err
}

func GoPublish(videoList []dao.Video, userID uint64) error {
	keyPublish := "Publish:" + strconv.FormatUint(userID, 10)
	var listZ = make([]*redis.Z, 0, len(videoList))
	for _, video := range videoList {
		listZ = append(listZ, &redis.Z{float64(video.CreatedAt.UnixMilli()) / 1000, video.VideoID})
	}
	return global.GVAR_REDIS.ZAdd(global.GVAR_CONTEXT, keyPublish, listZ...).Err()
}

func GoCommentsOfVideo(commentList []dao.Comment, keyCommentsOfVideo string) error {
	var listZ = make([]*redis.Z, 0, len(commentList))
	for _, comment := range commentList {
		listZ = append(listZ, &redis.Z{float64(comment.CreatedAt.UnixMilli()) / 1000, comment.CommentID})
	}
	return global.GVAR_REDIS.ZAdd(global.GVAR_CONTEXT, keyCommentsOfVideo, listZ...).Err()
}

// 返回视频数
func GetFeedVideosAndAuthorsRedis(videoList *[]dao.Video, authors *[]dao.User, LatestTime int64, MaxNumVideo int) (int64, error) {
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, "feed").Result()
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		// "feed"不存在
		var allVideos []dao.Video
		if err = global.GVAR_DB.Find(&allVideos).Error; err != nil {
			return 0, err
		}
		if len(allVideos) == 0 {
			return 0, nil
		}
		if err = GoFeed(allVideos); err != nil {
			return 0, err
		}
	}
	// 初始化查询条件， Offset和Count用于分页
	op := redis.ZRangeBy{
		Min:    "0",                                                         // 最小分数
		Max:    strconv.FormatFloat(float64(LatestTime-2)/1000, 'f', 3, 64), // 最大分数
		Offset: 0,                                                           // 类似sql的limit, 表示开始偏移量
		Count:  int64(MaxNumVideo),                                          // 一次返回多少数据
	}

	vals, err := global.GVAR_REDIS.ZRevRangeByScore(global.GVAR_CONTEXT, "feed", &op).Result()
	if err != nil || len(vals) == 0 {
		return 0, err
	}
	*videoList = make([]dao.Video, 0, len(vals))
	authorIDList := make([]uint64, 0, len(vals))

	for _, videoIDStr := range vals {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		keyVideo := "Video:" + videoIDStr
		var video dao.Video
		n, err = global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyVideo).Result()
		if err != nil {
			return 0, err
		}
		if n <= 0 {
			// "video_id"不存在
			result := global.GVAR_DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
			if err = GoVideo(video); err != nil {
				return 0, err
			}
			if result.Error != nil || result.RowsAffected == 0 {
				return 0, errors.New("Get video fail")
			} else {
				authorIDList = append(authorIDList, video.AuthorID)
				*videoList = append(*videoList, video)
			}
			continue
		}
		if err = global.GVAR_REDIS.HGetAll(global.GVAR_CONTEXT, keyVideo).Scan(&video); err != nil {
			return 0, err
		} else {
			video.VideoID = videoID
			if err != nil {
				continue
			}
			timeUnixMilliStr, err := global.GVAR_REDIS.HGet(global.GVAR_CONTEXT, keyVideo, "created_at").Result()
			if err != nil {
				continue
			}
			timeUnixMilli, err := strconv.ParseInt(timeUnixMilliStr, 10, 64)
			if err != nil {
				continue
			}
			video.CreatedAt = time.UnixMilli(timeUnixMilli)
			authorIDList = append(authorIDList, video.AuthorID)
			*videoList = append(*videoList, video)
		}
	}
	err = GetUserListByUserIDs(authorIDList, authors)
	if err != nil {
		return 0, err
	} else {
		return int64(len(*videoList)), nil
	}
}

// PublishVideo 记录接收视频的属性并写入数据库
func PublishVideo(userID uint64, videoID uint64, videoName string, coverName string, title string) error {
	var video dao.Video
	video.VideoID = videoID
	video.Title = title
	video.PlayName = videoName
	video.CoverName = coverName
	//video.FavoriteCount = 0
	//video.CommentCount = 0
	video.AuthorID = userID
	//video.CreatedAt = time.Now().UnixMilli()

	err := global.GVAR_DB.Create(&video).Error
	return err
}

func PublishVideoRedis(userID uint64, videoID uint64, videoName string, coverName string, title string) error {
	keyPublish := fmt.Sprintf("Publish:%d", userID)
	videoIDStr := strconv.FormatUint(videoID, 10)
	Z := redis.Z{float64(time.Now().Unix()), videoIDStr}
	if err := global.GVAR_REDIS.ZAdd(global.GVAR_CONTEXT, keyPublish, &Z).Err(); err != nil {
		return err
	}
	keyVideo := fmt.Sprintf("Video:%d", videoID)
	if err := global.GVAR_REDIS.HSet(global.GVAR_CONTEXT, keyVideo, "author_id", userID, "play_name", videoName, "cover_name", coverName,
		"favorite_count", 0, "comment_count", 0, "title", title, "created_at", time.Now().UnixMilli()).Err(); err != nil {
		return err
	}

	Z = redis.Z{float64(time.Now().UnixMilli()) / 1000, videoID}
	err := global.GVAR_REDIS.ZAdd(global.GVAR_CONTEXT, "feed", &Z).Err()
	return err
}

// GetPublishedVideosAndAuthors 按要求拉取feed视频和其作者
func GetPublishedVideosAndAuthorsRedis(videoList *[]dao.Video, authors *[]dao.User, userID uint64) (int64, error) {
	keyPublish := "Publish:" + strconv.FormatUint(userID, 10)
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyPublish).Result()
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		// "publish userid"不存在
		result := global.GVAR_DB.Where("author_id = ?", userID).Find(videoList)
		if result.Error != nil {
			return 0, err
		} else if result.RowsAffected == 0 {
			return 0, nil
		} else {
			// 成功
			// 翻转
			numVideos := int(result.RowsAffected)
			for i, j := 0, numVideos-1; i < j; i, j = i+1, j-1 {
				(*videoList)[i], (*videoList)[j] = (*videoList)[j], (*videoList)[i]
			}
			if err = GoPublish(*videoList, userID); err != nil {
				return 0, err
			}
			authorIDList := make([]uint64, numVideos)
			for i, video := range *videoList {
				authorIDList[i] = video.AuthorID
			}
			err := GetUserListByUserIDs(authorIDList, authors)
			if err != nil {
				return 0, err
			} else {
				return result.RowsAffected, nil
			}
		}
	} else {
		vals, err := global.GVAR_REDIS.ZRevRange(global.GVAR_CONTEXT, keyPublish, 0, -1).Result()
		if err != nil {
			return 0, err
		}
		*videoList = make([]dao.Video, 0, len(vals))
		authorIDList := make([]uint64, 0, len(vals))

		for _, video_id := range vals {
			var video dao.Video
			n, err = global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, "Video:"+video_id).Result()
			if err != nil {
				return 0, err
			}
			if n <= 0 {
				// "video_id"不存在
				result := global.GVAR_DB.Where("video_id = ?", video_id).Limit(1).Find(&video)
				if err = GoVideo(video); err != nil {
					return 0, err
				}
				if result.Error != nil || result.RowsAffected == 0 {
					return 0, errors.New("Get video fail")
				} else {
					authorIDList = append(authorIDList, video.AuthorID)
					*videoList = append(*videoList, video)
				}
				continue
			}
			if err = global.GVAR_REDIS.HGetAll(global.GVAR_CONTEXT, "Video:"+video_id).Scan(&video); err != nil {
				return 0, err
			} else {
				video.VideoID, err = strconv.ParseUint(video_id, 10, 64)
				if err != nil {
					continue
				}
				timeUnixMilliStr, err := global.GVAR_REDIS.HGet(global.GVAR_CONTEXT, "Video:"+video_id, "created_at").Result()
				if err != nil {
					continue
				}
				timeUnixMilli, err := strconv.ParseInt(timeUnixMilliStr, 10, 64)
				if err != nil {
					continue
				}
				video.CreatedAt = time.UnixMilli(timeUnixMilli)
				authorIDList = append(authorIDList, video.AuthorID)
				*videoList = append(*videoList, video)
			}
		}
		err = GetUserListByUserIDs(authorIDList, authors)
		if err != nil {
			return 0, err
		} else {
			return int64(len(*videoList)), nil
		}
	}
}

// GetVideoListByIDs 给定视频ID列表得到对应的视频信息
func GetVideoListByIDsRedis(videoList *[]dao.Video, videoIDs []uint64) error {
	*videoList = make([]dao.Video, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		var video dao.Video
		keyVideo := "Video:" + strconv.FormatUint(videoID, 10)
		n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyVideo).Result()
		if err != nil {
			return err
		}
		if n <= 0 {
			result := global.GVAR_DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
			if result.Error != nil || result.RowsAffected == 0 {
				return errors.New("GetVideoListByIDsRedis fail")
			}
			if err = GoVideo(video); err != nil {
				return err
			}
			*videoList = append(*videoList, video)
		} else {
			if err = global.GVAR_REDIS.HGetAll(global.GVAR_CONTEXT, keyVideo).Scan(&video); err != nil {
				return errors.New("GetVideoListByIDsRedis fail")
			} else {
				video.VideoID = videoID
				timeUnixMilliStr, err := global.GVAR_REDIS.HGet(global.GVAR_CONTEXT, "Video:"+strconv.FormatUint(videoID, 10), "created_at").Result()
				if err != nil {
					continue
				}
				timeUnixMilli, err := strconv.ParseInt(timeUnixMilliStr, 10, 64)
				if err != nil {
					continue
				}
				video.CreatedAt = time.UnixMilli(timeUnixMilli)
				*videoList = append(*videoList, video)
			}
		}
	}
	return nil
}
