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

func GetFavoriteStatusFromRedis(userID, videoID uint64) (bool, error) {
	// 定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
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
	keys := []string{userFavoriteRedis}
	values := []interface{}{videoID, global.FAVORITE_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return false, errors.New("not found in cache")
	} else {
		return false, err
	}
}

func AddFavoriteVideoIDListByUserIDToRedis(userID uint64, favoriteList []model.Favorite) error {
	// 定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	// 使用 pipeline
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.ZAdd(global.CONTEXT, userFavoriteRedis, &redis.Z{Score: 2, Member: Header})
		// 增加点赞关系
		for _, each := range favoriteList {
			if each.IsFavorite {
				pipe.ZAdd(global.CONTEXT, userFavoriteRedis, &redis.Z{Score: 1, Member: each.VideoID})
			} else {
				pipe.ZAdd(global.CONTEXT, userFavoriteRedis, &redis.Z{Score: 0, Member: each.VideoID})
			}
		}
		//设置过期时间
		pipe.Expire(global.CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
		return nil
	})
	return err
}

func AddFavoriteForRedis(videoID, userID, authorID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	// 更新 userFavoriteRedis 缓存
	go func() {
		// 定义 key
		userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("ZAdd", KEYS[1], 1, ARGV[1])
					redis.call("Expire", KEYS[1], ARGV[2])
					return true
				end
				return false
			`)
		keys := []string{userFavoriteRedis}
		values := []interface{}{videoID, global.FAVORITE_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 userRedis 缓存
	go func() {
		// 定义 key
		userRedis := fmt.Sprintf(UserPattern, userID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "favorite_count", 1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{userRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 authorRedis 缓存
	go func() {
		// 定义 key
		authorRedis := fmt.Sprintf(UserPattern, authorID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "total_favorited", 1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{authorRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 videoRedis 缓存
	go func() {
		// 定义 key
		videoRedis := fmt.Sprintf(VideoPattern, videoID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "favorite_count", 1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{videoRedis}
		values := []interface{}{global.VIDEO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
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

func CancelFavoriteForRedis(videoID, userID, authorID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	// 更新 userFavoriteRedis 缓存
	go func() {
		// 定义 key
		userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("ZAdd", KEYS[1], 0, ARGV[1])
					redis.call("Expire", KEYS[1], ARGV[2])
					return true
				end
				return false
			`)
		keys := []string{userFavoriteRedis}
		values := []interface{}{videoID, global.FAVORITE_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 userRedis 缓存
	go func() {
		// 定义 key
		userRedis := fmt.Sprintf(UserPattern, userID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "favorite_count", -1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{userRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 authorRedis 缓存
	go func() {
		// 定义 key
		authorRedis := fmt.Sprintf(UserPattern, authorID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "total_favorited", -1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{authorRedis}
		values := []interface{}{global.USER_INFO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Bool()
		ch <- err
	}()

	// 更新 videoRedis 缓存
	go func() {
		// 定义 key
		videoRedis := fmt.Sprintf(VideoPattern, videoID)
		lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("HIncrBy", KEYS[1], "favorite_count", -1)
					redis.call("Expire", KEYS[1], ARGV[1])
					return true
				end
				return false
			`)
		keys := []string{videoRedis}
		values := []interface{}{global.VIDEO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
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

func GetFavoriteVideoIDListByUserIDFromRedis(userID uint64) ([]uint64, error) {
	// 定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	lua := redis.NewScript(`
			if redis.call("Exists", KEYS[1]) <= 0 then
				return false
			end
			redis.call("Expire", KEYS[1], ARGV[1])
			return redis.call("ZRangeByScore", KEYS[1], 1, 1)
			`)
	keys := []string{userFavoriteRedis}
	values := []interface{}{global.FAVORITE_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Uint64Slice()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return nil, errors.New("not found in cache")
	} else {
		return nil, err
	}
}

func GetFavoriteCountByVideoIDFromRedis(videoID uint64) (int64, error) {
	// 定义 key
	videoRedis := fmt.Sprintf(VideoPattern, videoID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("Expire", KEYS[1], ARGV[1])
					return redis.call("HGet", KEYS[1], "favorite_count")
				end
				return false
			`)
	keys := []string{videoRedis}
	values := []interface{}{global.VIDEO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Int64()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return 0, errors.New("not found in cache")
	} else {
		return 0, err
	}
}

func AddFavoriteCountByVideoIDToRedis(videoID uint64, favoriteCount int64) error {
	// 定义 key
	videoRedis := fmt.Sprintf(VideoPattern, videoID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					redis.call("Expire", KEYS[1], ARGV[2])
					return redis.call("HSet", KEYS[1], "favorite_count", ARGV[1])
				end
				return false
			`)
	keys := []string{videoRedis}
	values := []interface{}{favoriteCount, global.VIDEO_EXPIRE.Seconds() + math.Floor(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())}
	err := lua.Run(global.CONTEXT, global.REDIS, keys, values).Err()
	if err == nil || err == redis.Nil {
		return nil
	} else {
		return err
	}
}

func GetFavoriteCountListByVideoIDListFromRedis(videoIDList []uint64) (favoriteCountList []int64, notInCache []uint64, err error) {
	// 定义 key
	userNum := len(videoIDList)
	favoriteCountList = make([]int64, 0, userNum)
	notInCache = make([]uint64, 0, userNum)
	for _, each := range videoIDList {
		favoriteCount, err2 := GetFavoriteCountByVideoIDFromRedis(each)
		if err2 != nil && err2.Error() != "not found in cache" {
			return nil, nil, err2
		} else if err2 == nil {
			favoriteCountList = append(favoriteCountList, favoriteCount)
		} else {
			err = err2
			favoriteCountList = append(favoriteCountList, -1)
			notInCache = append(notInCache, each)
		}
	}
	return
}

func AddFavoriteCountListByUVideoIDListToCache(videoList []VideoFavoriteCountAPI) error {
	// 使用 pipeline
	for _, each := range videoList {
		if err := AddFavoriteCountByVideoIDToRedis(each.VideoID, each.FavoriteCount); err != nil {
			return err
		}
	}
	return nil
}
