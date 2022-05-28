package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
)

func DeleteFollowFromCache(followerID, celebrityID uint64) error {
	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
	followerRedis := fmt.Sprintf(UserPattern, followerID)
	celebrityRedis := fmt.Sprintf(UserPattern, celebrityID)

	// 删除缓存
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(CTX, func(pipe redis.Pipeliner) error {
			// 删除关注关系
			pipe.Del(CTX, followerRelationRedis, celebrityRelationRedis)

			//删除redis user相关
			pipe.Del(CTX, followerRedis, celebrityRedis)
			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, followerRelationRedis, celebrityRelationRedis, followerRedis, celebrityRedis)
}

func GetFollowIDListByUserIDFromCache(followerID uint64) ([]uint64, error) {
	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)

	if result := global.GVAR_REDIS.Exists(CTX, followerRelationRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	celebrityIDStrList := global.GVAR_REDIS.SMembers(CTX, followerRelationRedis).Val()
	global.GVAR_REDIS.Expire(CTX, followerRelationRedis, global.FOLLOW_EXPIRE)
	celebrityIDList := make([]uint64, 0, len(celebrityIDStrList))
	for i := 0; i < len(celebrityIDStrList); i++ {
		if celebrityIDStrList[i] == HEADER {
			continue
		}
		celebrityID, err := strconv.ParseUint(celebrityIDStrList[i], 10, 64)
		if err != nil {
			return nil, errors.New("Wrong format conversion in cache")
		}
		celebrityIDList = append(celebrityIDList, celebrityID)
	}
	return celebrityIDList, nil
}

func AddFollowIDListByUserIDInCache(followerID uint64, celebrityIDList []uint64) error {
	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(CTX, func(pipe redis.Pipeliner) error {
			// 初始化
			pipe.SAdd(CTX, followerRelationRedis, HEADER)

			// 增加点赞关系
			for _, each := range celebrityIDList {
				pipe.SAdd(CTX, followerRelationRedis, each)
			}

			//设置过期时间
			pipe.Expire(CTX, followerRelationRedis, global.FOLLOW_EXPIRE)
			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, followerRelationRedis)
}

func GetFollowerIDListByUserIDFromCache(celebrityID uint64) ([]uint64, error) {
	//定义 key
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)

	if result := global.GVAR_REDIS.Exists(CTX, celebrityRelationRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	followerIDStrList := global.GVAR_REDIS.SMembers(CTX, celebrityRelationRedis).Val()
	global.GVAR_REDIS.Expire(CTX, celebrityRelationRedis, global.FOLLOW_EXPIRE)
	followerIDList := make([]uint64, 0, len(followerIDStrList))
	for i := 0; i < len(followerIDStrList); i++ {
		if followerIDStrList[i] == HEADER {
			continue
		}
		celebrityID, err := strconv.ParseUint(followerIDStrList[i], 10, 64)
		if err != nil {
			return nil, errors.New("Wrong format conversion in cache")
		}
		followerIDList = append(followerIDList, celebrityID)
	}
	return followerIDList, nil
}

func AddFollowerIDListByUserIDInCache(celebrityID uint64, followerIDList []uint64) error {
	//定义 key
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(CTX, func(pipe redis.Pipeliner) error {
			// 初始化
			pipe.SAdd(CTX, celebrityRelationRedis, HEADER)

			// 增加点赞关系
			for _, each := range followerIDList {
				pipe.SAdd(CTX, celebrityRelationRedis, each)
			}

			//设置过期时间
			pipe.Expire(CTX, celebrityRelationRedis, global.FOLLOW_EXPIRE)
			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, celebrityRelationRedis)
}
