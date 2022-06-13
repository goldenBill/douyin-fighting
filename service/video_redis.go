package service

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// GoPublish 将用户发表过的视频写入缓存中
func GoPublish(userID uint64, listZ ...*redis.Z) error {
	//定义 key
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, keyPublish, listZ...)
	pipe.Expire(global.CONTEXT, keyPublish, global.PUBLISH_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

// GoVideoList 将视频批量写入缓存
func GoVideoList(videoList []model.Video) error {
	pipe := global.REDIS.TxPipeline()
	for _, video := range videoList {
		keyVideo := fmt.Sprintf(VideoPattern, video.VideoID)
		pipe.HSet(global.CONTEXT, keyVideo, "title", video.Title, "play_name", video.PlayName, "cover_name", video.CoverName,
			"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount, "author_id", video.AuthorID, "created_at", video.CreatedAt.UnixMilli())
		pipe.Expire(global.CONTEXT, keyVideo, global.VIDEO_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
	}
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

// GoFeed 确保feed在缓存中
func GoFeed() error {
	n, err := global.REDIS.Exists(global.CONTEXT, "feed").Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// "feed"不存在
		var allVideos []model.Video
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

// PublishEvent 用户上视频的缓存操作
func PublishEvent(video model.Video, listZ ...*redis.Z) error {
	keyPublish := fmt.Sprintf(PublishPattern, video.AuthorID)
	keyVideo := fmt.Sprintf(VideoPattern, video.VideoID)
	keyEmpty := fmt.Sprintf(EmptyPattern, video.AuthorID)
	videoIDStr := strconv.FormatUint(video.VideoID, 10)
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, "feed", &redis.Z{Score: float64(video.CreatedAt.UnixMilli()) / 1000, Member: videoIDStr})
	pipe.ZAdd(global.CONTEXT, keyPublish, listZ...)
	pipe.Expire(global.CONTEXT, keyPublish, global.PUBLISH_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)

	pipe.HSet(global.CONTEXT, keyVideo, "author_id", video.AuthorID, "play_name", video.PlayName, "cover_name", video.CoverName,
		"favorite_count", video.FavoriteCount, "comment_count", video.CommentCount, "title", video.Title, "created_at", video.CreatedAt.UnixMilli())
	pipe.Expire(global.CONTEXT, keyVideo, global.VIDEO_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
	pipe.Del(global.CONTEXT, keyEmpty)
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

func GoCommentsOfVideo(commentList []model.Comment, keyCommentsOfVideo string) error {
	var listZ = make([]*redis.Z, 0, len(commentList))
	for _, comment := range commentList {
		listZ = append(listZ, &redis.Z{Score: float64(comment.CreatedAt.UnixMilli()) / 1000, Member: comment.CommentID})
	}
	pipe := global.REDIS.TxPipeline()
	pipe.ZAdd(global.CONTEXT, keyCommentsOfVideo, listZ...)
	pipe.Expire(global.CONTEXT, keyCommentsOfVideo, global.VIDEO_COMMENTS_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

func GetCommentCountOfVideo(videoID uint64) (int, error) {
	keyVideo := fmt.Sprintf(VideoPattern, videoID)
	lua := redis.NewScript(`
				local key = KEYS[1]
				local expire_time = ARGV[1]
				if redis.call("Exists", key) > 0 then
					redis.call("Expire", key, expire_time)
					return redis.call("HGet", key, "comment_count")
				end
				return -1
			`)
	keys := []string{keyVideo}
	values := []interface{}{global.VIDEO_COMMENTS_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	numComments, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Int()
	if err != nil {
		return 0, err
	}
	return numComments, nil
}

func SetUserPublishEmpty(userID uint64) error {
	keyEmpty := fmt.Sprintf(EmptyPattern, userID)
	return global.REDIS.Set(global.CONTEXT, keyEmpty, "1", global.EMPTY_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second).Err()
}
