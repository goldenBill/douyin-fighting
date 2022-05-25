package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

//userid是否关注touserid
func IsFollow(userId, toUserId uint64) bool {
	var follow dao.Follow

	global.GVAR_DB.Where(&dao.Follow{
		UserID:   userId,
		FollowId: toUserId,
	}).First(&follow)
	if follow.ID == 0 {
		return false
	} else {
		return true
	}
}

//关注
func AddFollow(userId, toUserId uint64) {
	follow := dao.Follow{
		UserID:   userId,
		FollowId: toUserId,
	}
	global.GVAR_DB.Create(&follow)
}

//取消关注
func RemoveFollow(userId, toUserId uint64) {
	var follow dao.Follow

	global.GVAR_DB.Where(&dao.Follow{
		UserID:   userId,
		FollowId: toUserId,
	}).First(&follow)

	global.GVAR_DB.Where(&dao.Follow{
		ID: follow.ID,
	}).Delete(&follow)
}

//获取所有关注者
func GetFollow(userId uint64) []dao.Follow {
	var follows []dao.Follow
	global.GVAR_DB.Where(&dao.Follow{
		UserID: userId,
	}).Find(&follows)
	return follows
}

//获取所有粉丝
func GetFans(userId uint64) []dao.Follow {
	var follows []dao.Follow
	global.GVAR_DB.Where(&dao.Follow{
		FollowId: userId,
	}).Find(&follows)
	return follows
}
