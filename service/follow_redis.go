package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"math"
	"math/rand"
	"time"
)

func GetFollowStatusFromRedis(followerID, celebrityID uint64) (bool, error) {
	// 定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
	lua := redis.NewScript(`
			if redis.call("Exists", KEYS[1]) <= 0 then
				return false
			end
			redis.call("Expire", KEYS[1], ARGV[2])
			local tmp = redis.call("ZScore", KEYS[1], ARGV[1])
			if not tmp then
				return {err = "no tracking information"}
			end
			return tmp
			`)
	keys := []string{followerRelationRedis}
	values := []interface{}{celebrityID, global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return false, errors.New("not found in cache")
	} else {
		return false, err
	}
}

func AddFollowIDListByUserIDToRedis(followerID uint64, celebrityList []model.Follow) error {
	// 定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
	// 使用 pipeline
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.ZAdd(global.CONTEXT, followerRelationRedis, &redis.Z{Score: 2, Member: Header})
		// 增加点赞关系
		for _, each := range celebrityList {
			if each.IsFollow {
				pipe.ZAdd(global.CONTEXT, followerRelationRedis, &redis.Z{Score: 1, Member: each.CelebrityID})
			} else {
				pipe.ZAdd(global.CONTEXT, followerRelationRedis, &redis.Z{Score: 0, Member: each.CelebrityID})
			}
		}
		// 设置过期时间
		pipe.Expire(global.CONTEXT, followerRelationRedis, global.FOLLOW_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
		return nil
	})
	return err
}

func AddFollowForRedis(followerID, celebrityID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	// 更新 followerRelationRedis 缓存
	go func() {
		// 定义 key
		followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("ZAdd", KEYS[1], 1, ARGV[1])
					redis.call("Expire", KEYS[1], ARGV[2])
					return true
				end
				return false
			`)
		keys := []string{followerRelationRedis}
		values := []interface{}{celebrityID, global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 celebrityRelationRedis 缓存
	go func() {
		// 定义 key
		celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("ZAdd", KEYS[1], 1, ARGV[1])
					redis.call("Expire", KEYS[1], ARGV[2])
					return true
				end
				return false
			`)
		keys := []string{celebrityRelationRedis}
		values := []interface{}{followerID, global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 followerRedis 缓存
	go func() {
		// 定义 key
		followerRedis := fmt.Sprintf(UserPattern, followerID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "follow_count", 1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{followerRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 celebrityRedis 缓存
	go func() {
		// 定义 key
		celebrityRedis := fmt.Sprintf(UserPattern, celebrityID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "follower_count", 1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{celebrityRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	var err error
	for i := 0; i < 4; i++ {
		errTmp := <-ch
		if errTmp != nil && errTmp != redis.Nil {
			err = errTmp
		}
	}
	return err
}

func CancelFollowForRedis(followerID, celebrityID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	// 更新 followerRelationRedis 缓存
	go func() {
		// 定义 key
		followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("ZAdd", KEYS[1], 0, ARGV[1])
					redis.call("Expire", KEYS[1], ARGV[2])
					return true
				end
				return false
			`)
		keys := []string{followerRelationRedis}
		values := []interface{}{celebrityID, global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 celebrityRelationRedis 缓存
	go func() {
		// 定义 key
		celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("ZAdd", KEYS[1], 0, ARGV[1])
					redis.call("Expire", KEYS[1], ARGV[2])
					return true
				end
				return false
			`)
		keys := []string{celebrityRelationRedis}
		values := []interface{}{followerID, global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 followerRedis 缓存
	go func() {
		// 定义 key
		followerRedis := fmt.Sprintf(UserPattern, followerID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "follow_count", -1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{followerRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 celebrityRedis 缓存
	go func() {
		// 定义 key
		celebrityRedis := fmt.Sprintf(UserPattern, celebrityID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "follower_count", -1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{celebrityRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	var err error
	for i := 0; i < 4; i++ {
		errTmp := <-ch
		if errTmp != nil && errTmp != redis.Nil {
			err = errTmp
		}
	}
	return err
}

func GetFollowIDListByUserIDFromRedis(followerID uint64) ([]uint64, error) {
	// 定义 key
	followerRelationRedis := fmt.Sprintf(FollowerPattern, followerID)
	lua := redis.NewScript(`
			if redis.call("Exists", KEYS[1]) <= 0 then
				return false
			end
			redis.call("Expire", KEYS[1], ARGV[1])
			return redis.call("ZRangeByScore", KEYS[1], 1, 1)
			`)
	keys := []string{followerRelationRedis}
	values := []interface{}{global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Uint64Slice()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return nil, errors.New("not found in cache")
	} else {
		return nil, err
	}
}

func GetFollowerIDListByUserIDFromRedis(celebrityID uint64) ([]uint64, error) {
	// 定义 key
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
	lua := redis.NewScript(`
			if redis.call("Exists", KEYS[1]) <= 0 then
				return false
			end
			redis.call("Expire", KEYS[1], ARGV[1])
			return redis.call("ZRangeByScore", KEYS[1], 1, 1)
			`)
	keys := []string{celebrityRelationRedis}
	values := []interface{}{global.FOLLOW_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Uint64Slice()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return nil, errors.New("not found in cache")
	} else {
		return nil, err
	}
}

func AddFollowerIDListByUserIDToRedis(celebrityID uint64, followerList []model.Follow) error {
	// 定义 key
	celebrityRelationRedis := fmt.Sprintf(CelebrityPattern, celebrityID)
	// 使用 pipeline
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		//初始化
		pipe.ZAdd(global.CONTEXT, celebrityRelationRedis, &redis.Z{Score: 2, Member: Header})
		// 增加点赞关系
		for _, each := range followerList {
			if each.IsFollow {
				pipe.ZAdd(global.CONTEXT, celebrityRelationRedis, &redis.Z{Score: 1, Member: each.FollowerID})
			} else {
				pipe.ZAdd(global.CONTEXT, celebrityRelationRedis, &redis.Z{Score: 0, Member: each.FollowerID})
			}
		}
		//设置过期时间
		pipe.Expire(global.CONTEXT, celebrityRelationRedis, global.FOLLOW_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
		return nil
	})
	return err
}
