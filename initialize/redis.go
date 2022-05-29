package initialize

import (
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
)

func InitRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})
	global.GVAR_REDIS = rdb
}
