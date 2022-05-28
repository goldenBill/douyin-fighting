package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
)

var (
	CTX                   context.Context = context.Background()
	UserPattern           string          = "user:%d"
	UserFavoritePattern   string          = "favorite:%d"
	VideoFavoritedPattern string          = "favorited:%d"
	CelebrityPattern      string          = "celebrity:%d"
	FollowerPattern       string          = "follower:%d"
)

func Retry(fn func(*redis.Tx) error, keys ...string) error {
	// Retry if the key has been changed.
	for i := 0; i < global.GVAR_MAX_RETRIES; i++ {
		err := global.GVAR_REDIS.Watch(CTX, fn, keys...)
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
