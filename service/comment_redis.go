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

// AddCommentInRedis 添加评论的redis相关操作
func AddCommentInRedis(comment *model.Comment) error {
	//定义 key
	keyCommentsOfVideo := fmt.Sprintf(VideoCommentsPattern, comment.VideoID)
	keyComment := fmt.Sprintf(CommentPattern, comment.CommentID)
	keyVideo := fmt.Sprintf(VideoPattern, comment.VideoID)
	// 判断keyCommentsOfVideo是否存在 存在则加入comment
	lua := redis.NewScript(`
				local key = KEYS[1]
				local score = ARGV[1]
				local comment_id = ARGV[2]
				local expire_time = ARGV[3]
				if redis.call("Exists", key) > 0 then
					redis.call("ZAdd", key, score, comment_id)
					redis.call("Expire", key, expire_time)
					return 1
				end
				return 0
			`)
	keys := []string{keyCommentsOfVideo}
	values := []interface{}{float64(comment.CreatedAt.UnixMilli()) / 1000, comment.CommentID,
		global.VIDEO_COMMENTS_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	if err != nil {
		return err
	}
	// 判断keyVideo是否存在，存在则comment_count加1
	lua = redis.NewScript(`
				local key = KEYS[1]
				local expire_time = ARGV[1]
				if redis.call("Exists", key) > 0 then
					redis.call("HIncrBy", key, "comment_count", 1)
					redis.call("Expire", key, expire_time)
					return 1
				end
				return 0
			`)
	keys = []string{keyVideo}
	values = []interface{}{global.COMMENT_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	_, err = lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	if err != nil {
		return err
	}
	// 加入comment，无需判断key是否存在
	userIDStr := strconv.FormatUint(comment.UserID, 10)
	videoIDStr := strconv.FormatUint(comment.VideoID, 10)
	pipe := global.REDIS.TxPipeline()
	pipe.Expire(global.CONTEXT, keyComment, global.COMMENT_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
	pipe.HSet(global.CONTEXT, keyComment, "content", comment.Content, "user_id", userIDStr, "video_id", videoIDStr, "created_at", time.Now().UnixMilli())
	_, err = pipe.Exec(global.CONTEXT)
	return err
}

// DeleteCommentInRedis 删除评论的redis相关操作
func DeleteCommentInRedis(videoID uint64, commentID uint64) error {
	//定义 key
	keyCommentsOfVideo := fmt.Sprintf(VideoCommentsPattern, videoID)
	keyComment := fmt.Sprintf(CommentPattern, commentID)
	keyVideo := fmt.Sprintf(VideoPattern, videoID)
	CommentIDStr := strconv.FormatUint(commentID, 10)
	// 判断keyCommentsOfVideo是否存在 存在则从有序集合中移除comment
	lua := redis.NewScript(`
				local key = KEYS[1]
				local comment_id = ARGV[1]
				local expire_time = ARGV[2]
				if redis.call("Exists", key) > 0 then
					redis.call("ZRem", key, comment_id)
					redis.call("Expire", key, expire_time)
					return 1
				end
				return 0
			`)
	keys := []string{keyCommentsOfVideo}
	values := []interface{}{CommentIDStr, global.VIDEO_COMMENTS_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	if err != nil {
		return err
	}
	// 判断keyVideo是否存在，存在则comment_count减1
	lua = redis.NewScript(`
				local key = KEYS[1]
				local expire_time = ARGV[1]
				if redis.call("Exists", key) > 0 then
					redis.call("HIncrBy", key, "comment_count", -1)
					redis.call("Expire", key, expire_time)
					return 1
				end
				return 0
			`)
	keys = []string{keyVideo}
	values = []interface{}{global.COMMENT_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	_, err = lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	// 删除comment，无需判断key是否存在
	return global.REDIS.Del(global.CONTEXT, keyComment).Err()
}

// GoComment 函数用来将给定comment写入redis，若已在redis中则什么都不做
func GoComment(comment model.Comment) error {
	keyComment := fmt.Sprintf(CommentPattern, comment.CommentID)
	lua := redis.NewScript(`
				local key = KEYS[1]
				local video_id = ARGV[1]
				local user_id = ARGV[2]
				local content = ARGV[3]
				local created_at = ARGV[4]
				local expire_time = ARGV[5]
				if redis.call("Exists", key) <= 0 then
					redis.call("HSet", key, "video_id", video_id, "user_id", user_id, "content", content, "created_at", created_at)
					redis.call("Expire", key, expire_time)
					return 1
				end
				return 0
			`)
	keys := []string{keyComment}
	values := []interface{}{comment.VideoID, comment.UserID, comment.Content, comment.CreatedAt.UnixMilli(),
		global.COMMENT_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	return err
}
