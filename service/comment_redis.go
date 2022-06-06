package service

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"strconv"
	"time"
)

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
	vals := []interface{}{float64(comment.CreatedAt.UnixMilli()) / 1000, comment.CommentID, global.VIDEO_COMMENTS_EXPIRE.Seconds()}
	_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
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
	vals = []interface{}{global.COMMENT_EXPIRE.Seconds()}
	_, err = lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
	if err != nil {
		return err
	}
	// 加入comment，无需判断key是否存在
	userIDStr := strconv.FormatUint(comment.UserID, 10)
	videoIDStr := strconv.FormatUint(comment.VideoID, 10)
	pipe := global.REDIS.TxPipeline()
	pipe.Expire(global.CONTEXT, keyComment, global.COMMENT_EXPIRE)
	pipe.HSet(global.CONTEXT, keyComment, "content", comment.Content, "user_id", userIDStr, "video_id", videoIDStr, "created_at", time.Now().UnixMilli())
	_, err = pipe.Exec(global.CONTEXT)
	return err
}

func DeleteCommentInRedis(videoID uint64, commentID uint64) error {
	//定义 key
	keyCommentsOfVideo := fmt.Sprintf(VideoCommentsPattern, videoID)
	keyComment := fmt.Sprintf(CommentPattern, commentID)
	keyVideo := fmt.Sprintf(VideoPattern, videoID)
	CommentIDStr := strconv.FormatUint(commentID, 10)
	// 判断keyCommentsOfVideo是否存在 存在则加入comment
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
	vals := []interface{}{CommentIDStr, global.VIDEO_COMMENTS_EXPIRE.Seconds()}
	_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
	if err != nil {
		return err
	}
	// 判断keyVideo是否存在，存在则comment_count加1
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
	vals = []interface{}{global.COMMENT_EXPIRE.Seconds()}
	_, err = lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
	// 删除comment，无需判断key是否存在
	return global.REDIS.Del(global.CONTEXT, keyComment).Err()
}

func GoComment(comment model.Comment) error {
	keyComment := fmt.Sprintf(CommentPattern, comment.CommentID)
	lua := redis.NewScript(`
				local key = KEYS[1]
				local video_id = ARGV[1]
				local user_id = ARGV[2]
				local content = ARGV[3]
				local created_at = ARGV[4]
				local expire_time = ARGV[5]
				if redis.call("Exists", key) > 0 then
					redis.call("HSet", key, "video_id", video_id, "user_id", user_id, "content", content, "created_at", created_at)
					redis.call("Expire", key, expire_time)
					return 1
				end
				return 0
			`)
	keys := []string{keyComment}
	vals := []interface{}{comment.VideoID, comment.UserID, comment.Content, comment.CreatedAt.UnixMilli(), global.COMMENT_EXPIRE.Seconds()}
	_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
	return err
}
