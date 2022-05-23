package model

import (
	"log"
	initialize "simple-demo/initialize"
)

type Follow struct {
	Id       int64 `json:"id"`
	UserId   int64 `json:"user_id"`
	FollowId int64 `json:"follow_id"`
}

//获取所有关注者
func GetFollow(user User) []Follow {
	var follows []Follow
	initialize.GLOBAL_DB.Where(&Follow{
		UserId: user.Id,
	}).Find(&follows)
	return follows
}

//获取所有粉丝
func GetFans(user User) []Follow {
	var follows []Follow
	initialize.GLOBAL_DB.Where(&Follow{
		FollowId: user.Id,
	}).Find(&follows)
	return follows
}

//userid是否关注touserid
func IsFollow(userId, toUserId int64) bool {
	var follow Follow
	initialize.GLOBAL_DB.Where(&Follow{
		UserId:   userId,
		FollowId: toUserId,
	}).First(&follow)
	log.Println("followid ", follow.Id)
	if follow.Id == 0 {
		return false
	} else {
		return true
	}
}

//关注
func AddFollow(userId, toUserId int64) {
	follow := Follow{
		UserId:   userId,
		FollowId: toUserId,
	}
	initialize.GLOBAL_DB.Create(&follow)
}

//取消关注
func RemoveFollow(userId, toUserId int64) {
	var follow Follow
	initialize.GLOBAL_DB.Where(&Follow{
		UserId:   userId,
		FollowId: toUserId,
	}).First(&follow)
	initialize.GLOBAL_DB.Where(&Follow{
		Id: follow.Id,
	}).Delete(&follow)
}
