package service

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"math/rand"
	"strconv"
	"time"
)

func GetUserInfoByUserIDFromRedis(userID uint64) (*model.User, error) {
	// 定义 key
	userRedis := fmt.Sprintf(UserPattern, userID)

	var user model.User
	if result := global.REDIS.Exists(global.CONTEXT, userRedis).Val(); result <= 0 {
		return nil, errors.New("not found in cache")
	}
	// 使用 pipeline
	cmds, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		pipe.HGetAll(global.CONTEXT, userRedis)
		pipe.HGet(global.CONTEXT, userRedis, "created_at").Val()
		// 设置过期时间
		pipe.Expire(global.CONTEXT, userRedis, global.USER_INFO_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = cmds[0].(*redis.StringStringMapCmd).Scan(&user); err != nil {
		return nil, err
	}
	timeUnixMilliStr := cmds[1].(*redis.StringCmd).Val()
	timeUnixMilli, _ := strconv.ParseInt(timeUnixMilliStr, 10, 64)
	user.CreatedAt = time.UnixMilli(timeUnixMilli)
	return &user, nil
}

func AddUserInfoByUserIDFromCacheToRedis(user *model.User) error {
	// 定义 key
	userRedis := fmt.Sprintf(UserPattern, user.UserID)

	// 使用 pipeline
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		pipe.HSet(global.CONTEXT, userRedis, "user_id", user.UserID)
		pipe.HSet(global.CONTEXT, userRedis, "name", user.Name)
		pipe.HSet(global.CONTEXT, userRedis, "password", user.Password)
		pipe.HSet(global.CONTEXT, userRedis, "follow_count", user.FollowerCount)
		pipe.HSet(global.CONTEXT, userRedis, "follower_count", user.FollowerCount)
		pipe.HSet(global.CONTEXT, userRedis, "total_favorited", user.TotalFavorited)
		pipe.HSet(global.CONTEXT, userRedis, "favorite_count", user.FavoriteCount)
		pipe.HSet(global.CONTEXT, userRedis, "created_at", user.CreatedAt.UnixMilli())
		// 设置过期时间
		pipe.Expire(global.CONTEXT, userRedis, global.USER_INFO_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
		return nil
	})
	return err
}

func GetUserListByUserIDListFromRedis(userIDList []uint64) (userList []model.User, notInCache []uint64, err error) {
	// 定义 key
	userNum := len(userIDList)
	userList = make([]model.User, 0, userNum)
	notInCache = make([]uint64, 0, userNum)
	for _, each := range userIDList {
		user, err2 := GetUserInfoByUserIDFromRedis(each)
		if err2 != nil && err2.Error() != "not found in cache" {
			return nil, nil, err2
		} else if err2 == nil {
			userList = append(userList, *user)
		} else {
			err = err2
			userList = append(userList, model.User{UserID: each})
			notInCache = append(notInCache, each)
		}
	}
	return
}

func AddUserListByUserIDListsToRedis(userList []model.User) error {
	// 使用 pipeline
	_, err := global.REDIS.TxPipelined(global.CONTEXT, func(pipe redis.Pipeliner) error {
		for _, each := range userList {
			// 定义 key
			userRedis := fmt.Sprintf(UserPattern, each.UserID)

			pipe.HSet(global.CONTEXT, userRedis, "user_id", each.UserID)
			pipe.HSet(global.CONTEXT, userRedis, "name", each.Name)
			pipe.HSet(global.CONTEXT, userRedis, "password", each.Password)
			pipe.HSet(global.CONTEXT, userRedis, "follow_count", each.FollowCount)
			pipe.HSet(global.CONTEXT, userRedis, "follower_count", each.FollowerCount)
			pipe.HSet(global.CONTEXT, userRedis, "total_favorited", each.TotalFavorited)
			pipe.HSet(global.CONTEXT, userRedis, "favorite_count", each.FavoriteCount)
			pipe.HSet(global.CONTEXT, userRedis, "created_at", each.CreatedAt.UnixMilli())
			// 设置过期时间
			pipe.Expire(global.CONTEXT, userRedis, global.USER_INFO_EXPIRE+time.Duration(rand.Float64()*global.EXPIRE_TIME_JITTER.Seconds())*time.Second)
		}
		return nil
	})
	return err
}
