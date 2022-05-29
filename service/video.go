package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// GetFeedVideosAndAuthors 按要求拉取feed视频和其作者
func GetFeedVideosAndAuthors(videos *[]dao.Video, authors *[]dao.User, LatestTime int64, MaxNumVideo int) (int64, error) {
	result := global.GVAR_DB.Where("created_at < ?", time.UnixMilli(LatestTime)).Order("created_at DESC").Limit(MaxNumVideo).Find(videos)
	if result.Error != nil {
		// 访问数据库出错
		return 0, result.Error
	} else if result.RowsAffected == 0 {
		// 没有满足条件的视频
		return 0, nil
	} else {
		// 成功
		authorIDList := make([]uint64, result.RowsAffected)
		for i, video := range *videos {
			authorIDList[i] = video.AuthorID
		}

		err := GetUserListByUserIDs(authorIDList, authors)
		if err != nil {
			return 0, err
		} else {
			return result.RowsAffected, nil
		}
	}
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

func GoPublish(videoIDList []string, userID uint64) error {
	keyPublish := "Publish:" + strconv.FormatUint(userID, 10)
	err := global.GVAR_REDIS.LPush(global.GVAR_CONTEXT, keyPublish, videoIDList).Err()
	return err
}

func GoCommentsOfVideo(commentIDList []string, videoID uint64) error {
	keyCommentsOfVideo := "CommentsOfVideo:" + strconv.FormatUint(videoID, 10)
	err := global.GVAR_REDIS.LPush(global.GVAR_CONTEXT, keyCommentsOfVideo, commentIDList).Err()
	return err
}

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

	for _, video_id := range vals {
		var video dao.Video
		n, err = global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, "Video:"+video_id).Result()
		if err != nil {
			return 0, err
		}
		if n <= 0 {
			// "video_id"不存在
			result := global.GVAR_DB.Where("video_id = ?", video_id).Limit(1).Find(&video)
			err = GoVideo(video)
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
	key := fmt.Sprintf("Publish:%d", userID)
	err := global.GVAR_REDIS.LPush(global.GVAR_CONTEXT, key, videoID).Err()
	if err != nil {
		return err
	}
	key = fmt.Sprintf("Video:%d", videoID)
	if err = global.GVAR_REDIS.HSet(global.GVAR_CONTEXT, key, "author_id", userID, "play_name", videoName, "cover_name", coverName,
		"favorite_count", 0, "comment_count", 0, "title", title, "created_at", time.Now().UnixMilli()).Err(); err != nil {
		return err
	}

	Z := redis.Z{float64(time.Now().UnixMilli()) / 1000, videoID}
	err = global.GVAR_REDIS.ZAdd(global.GVAR_CONTEXT, "feed", &Z).Err()
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
			var videoIDList = make([]string, len(*videoList))
			for i, video := range *videoList {
				videoIDList[i] = strconv.FormatUint(video.VideoID, 10)
			}
			GoPublish(videoIDList, userID)
			authorIDList := make([]uint64, result.RowsAffected)
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
		vals, err := global.GVAR_REDIS.LRange(global.GVAR_CONTEXT, keyPublish, 0, -1).Result()
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
				err = GoVideo(video)
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

// GetPublishedVideosAndAuthors 按要求拉取feed视频和其作者
func GetPublishedVideosAndAuthors(videos *[]dao.Video, authors *[]dao.User, userID uint64) (int64, error) {
	result := global.GVAR_DB.Where("author_id = ?", userID).Find(videos)
	if result.Error != nil {
		// 访问数据库出错
		return 0, result.Error
	} else if result.RowsAffected == 0 {
		// 没有满足条件的视频
		return 0, nil
	} else {
		// 成功
		authorIDList := make([]uint64, result.RowsAffected)
		for i, video := range *videos {
			authorIDList[i] = video.AuthorID
		}
		err := GetUserListByUserIDs(authorIDList, authors)
		if err != nil {
			return 0, err
		} else {
			return result.RowsAffected, nil
		}
	}
}

// GetPublishedVideos 给定用户ID得到其发表过的视频
func GetPublishedVideos(videos *[]dao.Video, userID uint64) error {
	err := global.GVAR_DB.Where("author_id = ?", userID).Find(videos).Error
	return err
}

// GetVideoListByIDs 给定视频ID列表得到对应的视频信息
func GetVideoListByIDs(videoList *[]dao.Video, videoIDs []uint64) error {
	var uniqueVideoList []dao.Video
	result := global.GVAR_DB.Where("video_id in ?", videoIDs).Find(&uniqueVideoList)

	if result.Error != nil {
		return errors.New("query GetVideoListByIDs error")
	}
	// 针对查询结果建立映射关系
	mapVideoIDToVideo := make(map[uint64]dao.Video)
	*videoList = make([]dao.Video, len(videoIDs))
	for idx, video := range uniqueVideoList {
		mapVideoIDToVideo[video.VideoID] = uniqueVideoList[idx]
	}
	// 构造返回值
	for idx, videoID := range videoIDs {
		(*videoList)[idx] = mapVideoIDToVideo[videoID]
	}
	return nil
}

// FavoriteCountPlus 点赞数加1
func FavoriteCountPlus(videoID uint64) error {
	err := global.GVAR_DB.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
	return err
}

// FavoriteCountMinus 点赞数减1
func FavoriteCountMinus(videoID uint64) error {
	err := global.GVAR_DB.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count - 1")).Error
	return err
}

// CommentCountPlus 评论数加1
func CommentCountPlus(videoID uint64) error {
	err := global.GVAR_DB.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count + 1")).Error
	return err
}

// CommentCountMinus 评论数减1
func CommentCountMinus(videoID uint64) error {
	err := global.GVAR_DB.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count - 1")).Error
	return err
}

// 判断当前videoID是否存在
func IsVideoExist(videoID uint64) bool {
	var count int64
	global.GVAR_DB.Model(&dao.Video{}).Where("video_id = ?", videoID).Count(&count)
	return count != 0
}
