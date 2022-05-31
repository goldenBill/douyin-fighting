package service

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
)

type void struct{}

var member void

var (
	HEADER              string = ""
	UserPattern         string = "user:%d"
	UserFavoritePattern string = "favorite:%d"
	CelebrityPattern    string = "celebrity:%d"
	FollowerPattern     string = "follower:%d"
)

func Retry(fn func(*redis.Tx) error, keys ...string) error {
	// Retry if the key has been changed.
	for i := 0; i < global.MAX_RETRIES; i++ {
		err := global.REDIS.Watch(global.CONTEXT, fn, keys...)
		if err == nil {
			// Success.
			return nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		return err
	}
	return errors.New("更新Redis失败")
}
