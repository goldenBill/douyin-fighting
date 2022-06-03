package initialize

import (
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/service"
)

func Redis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})
	if _, err := rdb.Ping(global.CONTEXT).Result(); err != nil {
		panic(err.Error())
	}
	global.REDIS = rdb
	if err := service.GoFeed(); err != nil {
		panic(err.Error())
	}
}
