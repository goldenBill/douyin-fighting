package service

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
)

func GoPublishRedis(userID uint64, listZ ...*redis.Z) error {
	//定义 key
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, keyPublish, listZ...)
	pipe.Expire(global.CONTEXT, keyPublish, global.PUBLISH_EXPIRE)
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

func GoVideoList(videoList []dao.Video) error {
	pipe := global.REDIS.TxPipeline()
	for _, video := range videoList {
		keyVideo := fmt.Sprintf(VideoPattern, video.VideoID)
		pipe.HSet(global.CONTEXT, keyVideo, "title", video.Title, "play_name", video.PlayName, "cover_name", video.CoverName,
			"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount, "author_id", video.AuthorID, "created_at", video.CreatedAt.UnixMilli())
		pipe.Expire(global.CONTEXT, keyVideo, global.VIDEO_EXPIRE)
	}
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

func GoFeed() error {
	n, err := global.REDIS.Exists(global.CONTEXT, "feed").Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// "feed"不存在
		var allVideos []dao.Video
		if err := global.DB.Find(&allVideos).Error; err != nil {
			return err
		}
		if len(allVideos) == 0 {
			return nil
		}
		var listZ = make([]*redis.Z, 0, len(allVideos))
		for _, video := range allVideos {
			listZ = append(listZ, &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: video.VideoID})
		}
		return global.REDIS.ZAdd(global.CONTEXT, "feed", listZ...).Err()
	}
	return nil
}

func PublishEvent(keyPublish string, video dao.Video, listZ ...*redis.Z) error {
	keyVideo := fmt.Sprintf("Video:%d", video.VideoID)
	videoIDStr := strconv.FormatUint(video.VideoID, 10)
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, "feed", &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: videoIDStr})
	pipe.ZAdd(global.CONTEXT, keyPublish, listZ...)
	pipe.Expire(global.CONTEXT, keyPublish, global.PUBLISH_EXPIRE)
	pipe.Expire(global.CONTEXT, keyVideo, global.VIDEO_EXPIRE)
	pipe.HSet(global.CONTEXT, keyVideo, "author_id", video.AuthorID, "play_name", video.PlayName, "cover_name", video.CoverName,
		"favorite_count", video.FavoriteCount, "comment_count", 0, "title", video.Title, "created_at", video.CreatedAt.UnixMilli())
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

func GoCommentsOfVideo(commentList []dao.Comment, keyCommentsOfVideo string) error {
	var listZ = make([]*redis.Z, 0, len(commentList))
	for _, comment := range commentList {
		listZ = append(listZ, &redis.Z{Score: float64(comment.CreatedAt.UnixMilli()) / 1000, Member: comment.CommentID})
	}
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, keyCommentsOfVideo, listZ...)
	pipe.Expire(global.CONTEXT, keyCommentsOfVideo, global.VIDEO_COMMENTS_EXPIRE)
	_, err := pipe.Exec(global.CONTEXT)
	return err
}
