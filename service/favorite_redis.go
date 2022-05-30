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
	videoRedis := "Video:" + strconv.FormatUint(videoID, 10)

	// 删除缓存
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 删除点赞关系
		pipe.Del(global.CONTEXT, userFavoriteRedis)
		//删除redis video相关
		pipe.Del(global.CONTEXT, videoRedis)
		//删除redis user相关
		pipe.Del(global.CONTEXT, userRedis, authorRedis)
		return nil
	})
	return err
}

func UpdateFavoriteActionFromCache(videoID, userID, authorID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	userRedis := fmt.Sprintf(UserPattern, userID)
	authorRedis := fmt.Sprintf(UserPattern, authorID)
	videoRedis := "Video:" + strconv.FormatUint(videoID, 10)

	// 更新 userFavoriteRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, userFavoriteRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.SAdd(global.CONTEXT, userFavoriteRedis, videoID)
				pipe.Expire(global.CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, userFavoriteRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, userFavoriteRedis).Err()
		}
		ch <- err
	}()

	// 更新 userRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, userRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, userRedis, "favorite_count", 1)
				pipe.Expire(global.CONTEXT, userRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, userRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, userRedis).Err()
		}
		ch <- err
	}()

	// 更新 authorRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, authorRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, authorRedis, "total_favorited", 1)
				pipe.Expire(global.CONTEXT, authorRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, authorRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, authorRedis).Err()
		}
		ch <- err
	}()

	// 更新 videoRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, videoRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, videoRedis, "favorite_count", 1)
				pipe.Expire(global.CONTEXT, videoRedis, global.VIDEO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, videoRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, videoRedis).Err()
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

func UpdateCancelFavoriteFromCache(videoID, userID, authorID uint64) error {
	// 设置管道
	ch := make(chan error, 2)
	defer close(ch)

	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	userRedis := fmt.Sprintf(UserPattern, userID)
	authorRedis := fmt.Sprintf(UserPattern, authorID)
	videoRedis := "Video:" + strconv.FormatUint(videoID, 10)

	// 更新 userFavoriteRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, userFavoriteRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.SRem(global.CONTEXT, userFavoriteRedis, videoID)
				pipe.Expire(global.CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, userFavoriteRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, userFavoriteRedis).Err()
		}
		ch <- err
	}()

	// 更新 userRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, userRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, userRedis, "favorite_count", -1)
				pipe.Expire(global.CONTEXT, userRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, userRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, userRedis).Err()
		}
		ch <- err
	}()

	// 更新 authorRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, authorRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, authorRedis, "total_favorited", -1)
				pipe.Expire(global.CONTEXT, authorRedis, global.USER_INFO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, authorRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, authorRedis).Err()
		}
		ch <- err
	}()

	// 更新 videoRedis 缓存
	go func() {
		txf := func(tx *redis.Tx) error {
			if result := global.REDIS.Exists(global.CONTEXT, videoRedis).Val(); result <= 0 {
				return nil
			}
			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
				pipe.HIncrBy(global.CONTEXT, videoRedis, "favorite_count", -1)
				pipe.Expire(global.CONTEXT, videoRedis, global.VIDEO_EXPIRE)
				return nil
			})
			return err
		}
		err := Retry(txf, videoRedis)
		if err != nil {
			err = global.REDIS.Del(global.CONTEXT, videoRedis).Err()
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

func GetFavoriteListByUserIDFromCache(userID uint64) ([]uint64, error) {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)

	if result := global.REDIS.Exists(global.CONTEXT, userFavoriteRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	// Transactional function.
	cmds, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		pipe.SMembers(global.CONTEXT, userFavoriteRedis).Val()
		pipe.Expire(global.CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
		return nil
	})
	if err != nil {
		return nil, err
	}
	videoIDStrList := cmds[0].(*redis.StringSliceCmd).Val()
	videoIDList := make([]uint64, 0, len(videoIDStrList)-1)
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
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		// 初始化
		pipe.SAdd(global.CONTEXT, userFavoriteRedis, HEADER)
		// 增加点赞关系
		for _, each := range videoIDList {
			pipe.SAdd(global.CONTEXT, userFavoriteRedis, each)
		}
		//设置过期时间
		pipe.Expire(global.CONTEXT, userFavoriteRedis, global.FAVORITE_EXPIRE)
		return nil
	})
	return err
}
