package initialize

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
)

func Redis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		panic(err.Error())
	}
	global.GVAR_REDIS = rdb
}
