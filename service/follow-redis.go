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

	_, err := global.GVAR_REDIS.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
		// 删除关注关系
		pipe.Del(global.GVAR_CONTEXT, followerRelationRedis, celebrityRelationRedis)

		//删除redis user相关
		pipe.Del(global.GVAR_CONTEXT, followerRedis, celebrityRedis)
		return nil
	})
	return err
}

func GetFollowIDListByUserIDFromCache(followerID uint64) ([]uint64, error) {
	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)

	if result := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, followerRelationRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	// Transactional function.
	var celebrityIDStrList []string
	if _, err := global.GVAR_REDIS.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
		celebrityIDStrList = global.GVAR_REDIS.SMembers(global.GVAR_CONTEXT, followerRelationRedis).Val()
		global.GVAR_REDIS.Expire(global.GVAR_CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	}); err != nil {
		return nil, err
	}
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
	_, err := global.GVAR_REDIS.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.SAdd(global.GVAR_CONTEXT, followerRelationRedis, HEADER)
		// 增加点赞关系
		for _, each := range celebrityIDList {
			pipe.SAdd(global.GVAR_CONTEXT, followerRelationRedis, each)
		}
		//设置过期时间
		pipe.Expire(global.GVAR_CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	})
	return err
}

func GetFollowerIDListByUserIDFromCache(celebrityID uint64) ([]uint64, error) {
	//定义 key
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)

	if result := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, celebrityRelationRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	// Transactional function.
	var followerIDStrList []string
	if _, err := global.GVAR_REDIS.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
		followerIDStrList = pipe.SMembers(global.GVAR_CONTEXT, celebrityRelationRedis).Val()
		pipe.Expire(global.GVAR_CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	}); err != nil {
		return nil, err
	}
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
	_, err := global.GVAR_REDIS.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.SAdd(global.GVAR_CONTEXT, celebrityRelationRedis, HEADER)
		// 增加点赞关系
		for _, each := range followerIDList {
			pipe.SAdd(global.GVAR_CONTEXT, celebrityRelationRedis, each)
		}
		//设置过期时间
		pipe.Expire(global.GVAR_CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	})
	return err
}
