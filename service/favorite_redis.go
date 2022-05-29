package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
)

func DeleteFavoriteFromCache(videoID, userID, authorID uint64) error {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	userRedis := fmt.Sprintf(UserPattern, userID)
	authorRedis := fmt.Sprintf(UserPattern, authorID)

	// 删除缓存
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
			// 删除点赞关系
			pipe.Del(global.GVAR_CONTEXT, userFavoriteRedis)

			//删除redis video相关
			/* Add your code here*/

			//删除redis user相关
			pipe.Del(global.GVAR_CONTEXT, userRedis, authorRedis)
			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, userFavoriteRedis, userRedis, authorRedis)
}

func GetFavoriteListByUserIDFromCache(userID uint64) ([]uint64, error) {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)

	if result := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, userFavoriteRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	videoIDStrList := global.GVAR_REDIS.SMembers(global.GVAR_CONTEXT, userFavoriteRedis).Val()
	global.GVAR_REDIS.Expire(global.GVAR_CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
	videoIDList := make([]uint64, 0, len(videoIDStrList))
	for i := 0; i < len(videoIDStrList); i++ {
		if videoIDStrList[i] == HEADER {
			continue
		}
		videoID, err := strconv.ParseUint(videoIDStrList[i], 10, 64)
		if err != nil {
			return nil, errors.New("Wrong format conversion in cache")
		}
		videoIDList = append(videoIDList, videoID)
	}
	return videoIDList, nil
}

func AddFavoriteListByUserIDInCache(userID uint64, videoIDList []uint64) error {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
			// 初始化
			pipe.SAdd(global.GVAR_CONTEXT, userFavoriteRedis, HEADER)

			// 增加点赞关系
			for _, each := range videoIDList {
				pipe.SAdd(global.GVAR_CONTEXT, userFavoriteRedis, each)
			}

			//设置过期时间
			pipe.Expire(global.GVAR_CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, userFavoriteRedis)
}
