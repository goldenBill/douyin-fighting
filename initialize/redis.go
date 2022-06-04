package initialize

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/service"
)

func Redis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", global.CONFIG.RedisConfig.Host, global.CONFIG.RedisConfig.Port),
		Password: global.CONFIG.RedisConfig.Password,
		DB:       global.CONFIG.RedisConfig.DB,
		PoolSize: global.CONFIG.RedisConfig.PoolSize,
	})
	if _, err := rdb.Ping(global.CONTEXT).Result(); err != nil {
		panic(err.Error())
	}
	global.REDIS = rdb
	if err := service.GoFeed(); err != nil {
		panic(err.Error())
	}
}
