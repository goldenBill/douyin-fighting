package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
	"time"
)

func GetUserInfoByUserIDFromCache(userID uint64) (*dao.User, error) {
	//定义 key
	userRedis := fmt.Sprintf(UserPattern, userID)

	var user dao.User
	if result := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, userRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	// Scan all fields into the model.
	if err := global.GVAR_REDIS.HGetAll(global.GVAR_CONTEXT, userRedis).Scan(&user); err != nil {
		panic(err)
	}
	timeUnixMilliStr := global.GVAR_REDIS.HGet(global.GVAR_CONTEXT, userRedis, "created_at").Val()
	global.GVAR_REDIS.Expire(global.GVAR_CONTEXT, userRedis, global.USER_INFO_EXPIRE)
	timeUnixMilli, _ := strconv.ParseInt(timeUnixMilliStr, 10, 64)
	user.CreatedAt = time.UnixMilli(timeUnixMilli)
	return &user, nil
}

func AddUserInfoByUserIDFromCacheInCache(user *dao.User) error {
	//定义 key
	userRedis := fmt.Sprintf(UserPattern, user.UserID)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(global.GVAR_CONTEXT, func(pipe redis.Pipeliner) error {
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "user_id", user.UserID)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "name", user.Name)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "password", user.Password)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "follow_count", user.FollowerCount)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "follower_count", user.FollowerCount)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "total_favorited", user.TotalFavorited)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "favorite_count", user.FavoriteCount)
			pipe.HSet(global.GVAR_CONTEXT, userRedis, "created_at", user.CreatedAt.UnixMilli())

			//设置过期时间
			pipe.Expire(global.GVAR_CONTEXT, userRedis, global.USER_INFO_EXPIRE)
			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, userRedis)
}

func GetUserListByUserIDListFromCache(userIDList []uint64) (userList []dao.User, notInCache []uint64, err error) {
	//定义 key
	userNum := len(userIDList)
	userList = make([]dao.User, 0, userNum)
	notInCache = make([]uint64, 0, userNum)
	for _, each := range userIDList {
		user, err2 := GetUserInfoByUserIDFromCache(each)
		if err2 != nil && err2.Error() != "Not found in cache" {
			return nil, nil, err2
		} else if err2 == nil {
			userList = append(userList, *user)
		} else {
			err = err2
			userList = append(userList, dao.User{UserID: each})
			notInCache = append(notInCache, each)
		}
	}
	return
}

func AddUserListByUserIDListsFromCacheInCache(userList []dao.User) error {
	for _, each := range userList {
		if err := AddUserInfoByUserIDFromCacheInCache(&each); err != nil {
			return err
		}
	}
	return nil
}
