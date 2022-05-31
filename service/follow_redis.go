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
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 删除关注关系
		pipe.Del(global.CONTEXT, followerRelationRedis, celebrityRelationRedis)
		//删除redis user相关
		pipe.Del(global.CONTEXT, followerRedis, celebrityRedis)
		return nil
	})
	return err
}

func UpdateFollowActionFromCache(followerID, celebrityID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
	followerRedis := fmt.Sprintf(UserPattern, followerID)
	celebrityRedis := fmt.Sprintf(UserPattern, celebrityID)

	// 更新 followerRelationRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, followerRelationRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.SAdd(global.CONTEXT, followerRelationRedis, celebrityID)
				pipe.Expire(global.CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, followerRelationRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, followerRelationRedis).Err()
		}
		ch <- err
	}()

	// 更新 celebrityRelationRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, celebrityRelationRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.SAdd(global.CONTEXT, celebrityRelationRedis, followerID)
				pipe.Expire(global.CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, celebrityRelationRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, celebrityRelationRedis).Err()
		}
		ch <- err
	}()

	// 更新 followerRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, followerRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, followerRedis, "follow_count", 1)
				pipe.Expire(global.CONTEXT, followerRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, followerRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, followerRedis).Err()
		}
		ch <- err
	}()

	// 更新 celebrityRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, celebrityRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, celebrityRedis, "follower_count", 1)
				pipe.Expire(global.CONTEXT, celebrityRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, celebrityRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, celebrityRedis).Err()
		}
		ch <- err
	}()

	for i := 0; i < 4; i++ {
		err := <-ch
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateCancelFollowFromCache(followerID, celebrityID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
	followerRedis := fmt.Sprintf(UserPattern, followerID)
	celebrityRedis := fmt.Sprintf(UserPattern, celebrityID)

	// 更新 followerRelationRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, followerRelationRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.SRem(global.CONTEXT, followerRelationRedis, celebrityID)
				pipe.Expire(global.CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, followerRelationRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, followerRelationRedis).Err()
		}
		ch <- err
	}()

	// 更新 celebrityRelationRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, celebrityRelationRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.SRem(global.CONTEXT, celebrityRelationRedis, followerID)
				pipe.Expire(global.CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, celebrityRelationRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, celebrityRelationRedis).Err()
		}
		ch <- err
	}()

	// 更新 followerRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, followerRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, followerRedis, "follow_count", -1)
				pipe.Expire(global.CONTEXT, followerRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, followerRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, followerRedis).Err()
		}
		ch <- err
	}()

	// 更新 celebrityRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := tx.Exists(global.CONTEXT, celebrityRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, celebrityRedis, "follower_count", -1)
				pipe.Expire(global.CONTEXT, celebrityRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, celebrityRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, celebrityRedis).Err()
		}
		ch <- err
	}()

	for i := 0; i < 4; i++ {
		err := <-ch
		if err != nil {
			return err
		}
	}
	return nil
}

func GetFollowIDListByUserIDFromCache(followerID uint64) ([]uint64, error) {
	//定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)

	if result := global.REDIS.Exists(global.CONTEXT, followerRelationRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	// Transactional function.
	cmds, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		pipe.SMembers(global.CONTEXT, followerRelationRedis).Val()
		pipe.Expire(global.CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	})
	if err != nil {
		return nil, err
	}
	celebrityIDStrList := cmds[0].(*redis.StringSliceCmd).Val()
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
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.SAdd(global.CONTEXT, followerRelationRedis, HEADER)
		// 增加点赞关系
		for _, each := range celebrityIDList {
			pipe.SAdd(global.CONTEXT, followerRelationRedis, each)
		}
		//设置过期时间
		pipe.Expire(global.CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	})
	return err
}

func GetFollowerIDListByUserIDFromCache(celebrityID uint64) ([]uint64, error) {
	//定义 key
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)

	if result := global.REDIS.Exists(global.CONTEXT, celebrityRelationRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	// Transactional function.
	cmds, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		pipe.SMembers(global.CONTEXT, celebrityRelationRedis).Val()
		pipe.Expire(global.CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	})
	if err != nil {
		return nil, err
	}
	followerIDStrList := cmds[0].(*redis.StringSliceCmd).Val()
	followerIDList := make([]uint64, 0, len(followerIDStrList)-1)
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
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.SAdd(global.CONTEXT, celebrityRelationRedis, HEADER)
		// 增加点赞关系
		for _, each := range followerIDList {
			pipe.SAdd(global.CONTEXT, celebrityRelationRedis, each)
		}
		//设置过期时间
		pipe.Expire(global.CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE)
		return nil
	})
	return err
}
