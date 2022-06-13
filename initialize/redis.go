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
	// 检查 Redis 连通性
	if _, err := rdb.Ping(global.CONTEXT).Result(); err != nil {
		panic(err.Error())
	}
	global.REDIS = rdb
	// 主动查询 feed，导入缓存
	if err := service.GoFeed(); err != nil {
		panic(err.Error())
	}
}
