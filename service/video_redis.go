package service

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
)

func GoPublishRedis(userID uint64, listZ ...*redis.Z) error {
	//定义 key
	keyPublish := fmt.Sprintf(PublishPattern, userID)
	return global.REDIS.ZAdd(global.CONTEXT, keyPublish, listZ...).Err()
}
