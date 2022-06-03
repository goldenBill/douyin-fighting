package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

func GetFavoriteStatusFromRedis(userID, videoID uint64) (bool, error) {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	lua := redis.NewScript(`
			if redis.call("Exists", KEYS[1]) <= 0 then
				return false
			end
			local tmp = redis.call("ZScore", KEYS[1], ARGV[1])
			if not tmp then
				return {err = "No tracking information"}
			end
			redis.call("Expire", KEYS[1], ARGV[2])
			return tmp
			`)
	keys := []string{userFavoriteRedis}
	vals := []interface{}{videoID, global.FAVORITE_EXPIRE.Seconds()}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return false, errors.New("Not found in cache")
	} else {
		return false, err
	}
}

func AddFavoriteVideoIDListByUserIDToRedis(userID uint64, favoriteList []dao.Favorite) error {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	// Transactional function.
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		//初始化
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
		pipe.Expire(global.CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
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
		//定义 key
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
		vals := []interface{}{videoID, global.FAVORITE_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
		ch <- err
	}()

	// 更新 userRedis 缓存
	go func() {
		//定义 key
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
		vals := []interface{}{global.USER_INFO_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
		ch <- err
	}()

	// 更新 authorRedis 缓存
	go func() {
		//定义 key
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
		vals := []interface{}{global.USER_INFO_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
		ch <- err
	}()

	// 更新 videoRedis 缓存
	go func() {
		//定义 key
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
		vals := []interface{}{global.VIDEO_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
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
		//定义 key
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
		vals := []interface{}{videoID, global.FAVORITE_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
		ch <- err
	}()

	// 更新 userRedis 缓存
	go func() {
		//定义 key
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
		vals := []interface{}{global.USER_INFO_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
		ch <- err
	}()

	// 更新 authorRedis 缓存
	go func() {
		//定义 key
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
		vals := []interface{}{global.USER_INFO_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
		ch <- err
	}()

	// 更新 videoRedis 缓存
	go func() {
		//定义 key
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
		vals := []interface{}{global.VIDEO_EXPIRE.Seconds()}
		_, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Bool()
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
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	lua := redis.NewScript(`
			if redis.call("Exists", KEYS[1]) <= 0 then
				return false
			end
			redis.call("Expire", KEYS[1], ARGV[1])
			return redis.call("ZRangeByScore", KEYS[1], 1, 1)
			`)
	keys := []string{userFavoriteRedis}
	vals := []interface{}{global.FAVORITE_EXPIRE.Seconds()}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Uint64Slice()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return nil, errors.New("Not found in cache")
	} else {
		return nil, err
	}
}

func GetFavoriteCountByVideoIDFromRedis(videoID uint64) (int64, error) {
	//定义 key
	videoRedis := fmt.Sprintf(VideoPattern, videoID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					return redis.call("HGet", KEYS[1], favorite_count)
				end
				return false
			`)
	keys := []string{videoRedis}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys).Int64()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return 0, errors.New("Not found in cache")
	} else {
		return 0, err
	}
}

func AddFavoriteCountByVideoIDToRedis(videoID uint64, favoriteCount int64) error {
	//定义 key
	videoRedis := fmt.Sprintf(VideoPattern, videoID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					return redis.call("HSet", KEYS[1], favorite_count, ARGV[1])
				end
				return false
			`)
	keys := []string{videoRedis}
	vals := []interface{}{favoriteCount}
	err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Err()
	if err == nil || err == redis.Nil {
		return nil
	} else {
		return err
	}
}

func GetFavoriteCountListByUVideoIDListFromRedis(videoIDList []uint64) (favoriteCountList []int64, notInCache []uint64, err error) {
	//定义 key
	userNum := len(videoIDList)
	favoriteCountList = make([]int64, 0, userNum)
	notInCache = make([]uint64, 0, userNum)
	for _, each := range videoIDList {
		favoriteCount, err2 := GetFavoriteCountByVideoIDFromRedis(each)
		if err2 != nil && err2.Error() != "Not found in cache" {
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

func AddFavoriteCountListByUVideoIDListToCache(videoList []dao.Video) error {
	// Transactional function.
	for _, each := range videoList {
		if err := AddFavoriteCountByVideoIDToRedis(each.VideoID, each.FavoriteCount); err != nil {
			return err
		}
	}
	return nil
}

func GetFavoriteCountByUserIDFromRedis(userID uint64) (int64, error) {
	//定义 key
	userRedis := fmt.Sprintf(UserPattern, userID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					return redis.call("HGet", KEYS[1], favorite_count)
				end
				return false
			`)
	keys := []string{userRedis}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys).Int64()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return 0, errors.New("Not found in cache")
	} else {
		return 0, err
	}
}

func AddFavoriteCountByUserIDToRedis(userID uint64, favoriteCount int64) error {
	//定义 key
	userRedis := fmt.Sprintf(UserPattern, userID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					return redis.call("HSet", KEYS[1], favorite_count, ARGV[1])
				end
				return false
			`)
	keys := []string{userRedis}
	vals := []interface{}{favoriteCount}
	err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Err()
	if err == nil || err == redis.Nil {
		return nil
	} else {
		return err
	}
}

func GetTotalFavoritedByUserIDFromRedis(userID uint64) (int64, error) {
	//定义 key
	userRedis := fmt.Sprintf(UserPattern, userID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					return redis.call("HGet", KEYS[1], total_favorited)
				end
				return false
			`)
	keys := []string{userRedis}
	result, err := lua.Run(global.CONTEXT, global.REDIS, keys).Int64()
	if err == nil {
		return result, nil
	} else if err == redis.Nil {
		return 0, errors.New("Not found in cache")
	} else {
		return 0, err
	}
}

func AddTotalFavoritedByUserIDToRedis(userID uint64, favoriteCount int64) error {
	//定义 key
	userRedis := fmt.Sprintf(UserPattern, userID)
	lua := redis.NewScript(`
				if redis.call("Exists", KEYS[1]) > 0 then
					return redis.call("HSet", KEYS[1], total_favorited, ARGV[1])
				end
				return false
			`)
	keys := []string{userRedis}
	vals := []interface{}{favoriteCount}
	err := lua.Run(global.CONTEXT, global.REDIS, keys, vals).Err()
	if err == nil || err == redis.Nil {
		return nil
	} else {
		return err
	}
}
