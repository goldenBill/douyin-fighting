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

func CheckVideo(videoID uint64) error {
	videoIDStr := strconv.FormatUint(videoID, 10)
	keyVideo := "Video:" + videoIDStr
	n, err := global.REDIS.Exists(global.CONTEXT, keyVideo).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// "video_id"不存在
		var video dao.Video
		result := global.DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("Redis出错或videoID不存在")
		}
		// 写redis Video:videoID
		err = GoVideo(&video)
	}
	return err
}

func GoFeed(allVideos []dao.Video) error {
	var listZ = make([]*redis.Z, 0, len(allVideos))
	for _, video := range allVideos {
		listZ = append(listZ, &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: video.VideoID})
	}
	return global.REDIS.ZAdd(global.CONTEXT, "feed", listZ...).Err()
}

func GoVideo(video *dao.Video) error {
	keyVideo := "Video:" + strconv.FormatUint(video.VideoID, 10)
	err := global.REDIS.HSet(global.CONTEXT, keyVideo, "title", video.Title, "play_name", video.PlayName, "cover_name", video.CoverName,
		"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount, "author_id", video.AuthorID, "created_at", video.CreatedAt.UnixMilli()).Err()
	return err
}

func PublishEvent(keyPublish string, video dao.Video, listZ ...*redis.Z) error {
	keyVideo := fmt.Sprintf("Video:%d", video.VideoID)
	videoIDStr := strconv.FormatUint(video.VideoID, 10)
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, "feed", &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: videoIDStr})
	pipe.ZAdd(global.CONTEXT, keyPublish, listZ...)
	pipe.HSet(global.CONTEXT, keyVideo, "author_id", video.AuthorID, "play_name", video.PlayName, "cover_name", video.CoverName,
		"favorite_count", 0, "comment_count", 0, "title", video.Title, "created_at", video.CreatedAt.UnixMilli())
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

func GoCommentsOfVideo(commentList []dao.Comment, keyCommentsOfVideo string) error {
	var listZ = make([]*redis.Z, 0, len(commentList))
	for _, comment := range commentList {
		listZ = append(listZ, &redis.Z{Score: float64(comment.CreatedAt.UnixMilli()) / 1000, Member: comment.CommentID})
	}
	return global.REDIS.ZAdd(global.CONTEXT, keyCommentsOfVideo, listZ...).Err()
}

// GetFeedVideosAndAuthorsRedis 返回视频数
func GetFeedVideosAndAuthorsRedis(videoList *[]dao.Video, authors *[]dao.User, LatestTime int64, MaxNumVideo int) (int64, error) {
	n, err := global.REDIS.Exists(global.CONTEXT, "feed").Result()
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		// "feed"不存在
		var allVideos []dao.Video
		if err = global.DB.Find(&allVideos).Error; err != nil {
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

	videoIDStrList, err := global.REDIS.ZRevRangeByScore(global.CONTEXT, "feed", &op).Result()
	numVideos := len(videoIDStrList)
	if err != nil || numVideos == 0 {
		return 0, err
	}

	*videoList = make([]dao.Video, 0, numVideos)
	authorIDList := make([]uint64, 0, numVideos)
	for _, videoIDStr := range videoIDStrList {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		keyVideo := fmt.Sprintf(VideoPattern, videoID)
		n, err = global.REDIS.Exists(global.CONTEXT, keyVideo).Result()
		if err != nil {
			continue
		}
		var video dao.Video
		if n <= 0 {
			// "video_id"不存在
			result := global.DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
			if err = GoVideo(&video); err != nil {
				return 0, err
			}
			if result.Error != nil || result.RowsAffected == 0 {
				return 0, errors.New("get video fail")
			}
			authorIDList = append(authorIDList, video.AuthorID)
			*videoList = append(*videoList, video)
			continue
		}
		// video_id 存在
		if err = global.REDIS.HGetAll(global.CONTEXT, keyVideo).Scan(&video); err != nil {
			continue
		}
		// redis中的Video:没有存video_ID
		video.VideoID = videoID
		// 字符串无法直接转化为time.time
		timeUnixMilliStr, err := global.REDIS.HGet(global.CONTEXT, keyVideo, "created_at").Result()
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
	if err = GetUserListByUserIDs(authorIDList, authors); err != nil {
		return 0, err
	}
	return int64(len(*videoList)), nil
}

func PublishVideoRedis(userID uint64, videoID uint64, videoName string, coverName string, title string) error {
	video := dao.Video{
		VideoID:   videoID,
		Title:     title,
		PlayName:  videoName,
		CoverName: coverName,
		//FavoriteCount : 0,
		//CommentCount : 0,
		AuthorID:  userID,
		CreatedAt: time.Now(),
	}
	if global.DB.Create(&video).Error != nil {
		return errors.New("video表插入失败")
	}
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyPublish).Result()
	if err != nil {
		return err
	}

	if n <= 0 {
		//	keyPublish不存在
		var videoList []dao.Video
		if err = global.DB.Where("author_id = ?", userID).Find(&videoList).Error; err != nil {
			return err
		}
		var listZ = make([]*redis.Z, 0, len(videoList))
		for _, video_ := range videoList {
			listZ = append(listZ, &redis.Z{Score: float64(video_.CreatedAt.UnixMilli()) / 1000, Member: video_.VideoID})
		}
		return PublishEvent(keyPublish, video, listZ...)
	}
	// keyPublish存在
	Z := redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: video.VideoID}
	return PublishEvent(keyPublish, video, &Z)
}

// GetPublishedVideosAndAuthorsRedis 按要求拉取feed视频和其作者
func GetPublishedVideosAndAuthorsRedis(videoList *[]dao.Video, authors *[]dao.User, userID uint64) (int, error) {
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyPublish).Result()
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		// "publish userid"不存在
		result := global.DB.Where("author_id = ?", userID).Find(videoList)
		numVideos := int(result.RowsAffected)
		if result.Error != nil || numVideos == 0 {
			return 0, err
		}
		var listZ = make([]*redis.Z, 0, numVideos)
		for _, video_ := range *videoList {
			listZ = append(listZ, &redis.Z{Score: float64(video_.CreatedAt.UnixMilli()) / 1000, Member: video_.VideoID})
		}
		// 写入publish：userid
		if err = GoPublishRedis(userID, listZ...); err != nil {
			return 0, err
		}

		authorIDList := make([]uint64, numVideos)
		for i, video := range *videoList {
			authorIDList[i] = video.AuthorID
		}
		if err = GetUserListByUserIDs(authorIDList, authors); err != nil {
			return 0, err
		}
		return numVideos, nil
	}
	// keyPublish存在
	videoIDStrList, err := global.REDIS.ZRevRange(global.CONTEXT, keyPublish, 0, -1).Result()
	numVideos := len(videoIDStrList)
	if err != nil {
		return 0, err
	}
	*videoList = make([]dao.Video, 0, numVideos)
	authorIDList := make([]uint64, 0, numVideos)

	for _, videoIDStr := range videoIDStrList {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		keyVideo := fmt.Sprintf(VideoPattern, videoID)
		n, err = global.REDIS.Exists(global.CONTEXT, keyVideo).Result()
		if err != nil {
			return 0, err
		}
		var video dao.Video
		if n <= 0 {
			// "video_id"不存在
			result := global.DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
			if err = GoVideo(&video); err != nil {
				return 0, err
			}
			if result.Error != nil || result.RowsAffected == 0 {
				return 0, errors.New("get video fail")
			}
			authorIDList = append(authorIDList, video.AuthorID)
			*videoList = append(*videoList, video)
			continue
		}
		// video_id存在 直接从redis中读入
		if err = global.REDIS.HGetAll(global.CONTEXT, keyVideo).Scan(&video); err != nil {
			return 0, err
		}
		video.VideoID = videoID
		timeUnixMilliStr, err := global.REDIS.HGet(global.CONTEXT, keyVideo, "created_at").Result()
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
	if err = GetUserListByUserIDs(authorIDList, authors); err != nil {
		return 0, err
	}
	return len(*videoList), nil
}

// GetVideoListByIDsRedis 给定视频ID列表得到对应的视频信息
func GetVideoListByIDsRedis(videoList *[]dao.Video, videoIDs []uint64) error {
	*videoList = make([]dao.Video, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		keyVideo := fmt.Sprintf(VideoPattern, videoID)
		n, err := global.REDIS.Exists(global.CONTEXT, keyVideo).Result()
		if err != nil {
			return err
		}
		var video dao.Video
		if n <= 0 {
			result := global.DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
			if result.Error != nil || result.RowsAffected == 0 {
				return errors.New("GetVideoListByIDsRedis fail")
			}
			if err = GoVideo(&video); err != nil {
				return err
			}
			*videoList = append(*videoList, video)
			continue
		}
		if err = global.REDIS.HGetAll(global.CONTEXT, keyVideo).Scan(&video); err != nil {
			return errors.New("GetVideoListByIDsRedis fail")
		}
		video.VideoID = videoID
		timeUnixMilliStr, err := global.REDIS.HGet(global.CONTEXT, keyVideo, "created_at").Result()
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
	return nil
}

func GetVideoIDListByUserID(userID uint64, videoIDList *[]uint64) error {
	keyPublish := fmt.Sprintf(VideoCommentsPattern, userID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyPublish).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// "publish userid"不存在
		var videoList []dao.Video
		result := global.DB.Where("author_id = ?", userID).Find(&videoList)
		if result.Error != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return nil
		}
		numVideos := int(result.RowsAffected)
		*videoIDList = make([]uint64, numVideos)
		var listZ = make([]*redis.Z, 0, numVideos)
		for i, videoID := range videoList {
			// 逆序 最新的放在前面
			(*videoIDList)[numVideos-i-1] = videoID.VideoID
		}
		for _, video := range videoList {
			listZ = append(listZ, &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: video.VideoID})
		}
		return GoPublishRedis(userID, listZ...)
	}
	// "publish userid"存在
	videoIDStrList, err := global.REDIS.ZRevRange(global.CONTEXT, keyPublish, 0, -1).Result()
	numVideos := len(videoIDStrList)
	*videoIDList = make([]uint64, 0, numVideos)
	for _, videoIDStr := range videoIDStrList {
		// 逆序 最新的放在前面
		videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
		if err != nil {
			continue
		}
		*videoIDList = append(*videoIDList, videoID)
	}
	return nil
}
