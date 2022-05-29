package initialize

import (
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
)

func Redis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})
	if _, err := rdb.Ping(global.GVAR_CONTEXT).Result(); err != nil {
		panic(err.Error())
	}
	global.GVAR_REDIS = rdb
}
