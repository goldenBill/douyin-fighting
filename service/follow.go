package service

import (
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
)

// GetFollowStatusForUpdate 获取关注状态，此处是针对 AddFollow 和 CancelFollow
func GetFollowStatusForUpdate(followerID, celebrityID uint64) (bool, error) {
	// 查询缓存
	followStatus, err := GetFollowStatusFromRedis(followerID, celebrityID)
	if err == nil {
		return followStatus, nil
	} else if err.Error() != "not found in cache" {
		return false, err
	}
	// 缓存不存在，查询数据库
	var followList []model.Follow
	if result := global.DB.Select("celebrity_id", "is_follow").Model(&model.Follow{}).
		Where("follower_id = ?", followerID).Find(&followList); result.Error != nil {
		return false, result.Error
	}
	// 更新缓存
	if err = AddFollowIDListByUserIDToRedis(followerID, followList); err != nil {
		return false, err
	}
	return GetFollowStatusFromRedis(followerID, celebrityID)
}

// GetFollowStatus 获取关注状态，此处是针对非更新操作
func GetFollowStatus(followerID, celebrityID uint64) (bool, error) {
	followStatus, err := GetFollowStatusForUpdate(followerID, celebrityID)
	if err == nil || err.Error() == "no tracking information" {
		return followStatus, nil
	}
	return false, err
}

// AddFollow 关注
func AddFollow(followerID, celebrityID uint64) error {
	// 获取当前关注状态
	if isFollow, err := GetFollowStatusForUpdate(followerID, celebrityID); err == nil {
		if isFollow {
			return nil
		}
		// 数据库有记录，修改数据库
		if err := global.DB.Model(&model.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
			Update("is_follow", true).Error; err != nil {
			return err
		}
	} else if err.Error() == "no tracking information" {
		var follow model.Follow
		// 数据库没有记录，写入数据库
		follow.FollowID, _ = global.ID_GENERATOR.NextID()
		follow.CelebrityID = celebrityID
		follow.FollowerID = followerID
		follow.IsFollow = true
		if err := global.DB.Create(&follow).Error; err != nil {
			return err
		}
	} else {
		return err
	}
	//更新缓存
	if err := AddFollowForRedis(followerID, celebrityID); err != nil {
		return err
	}
	return nil
}

// CancelFollow 取消关注
func CancelFollow(followerID, celebrityID uint64) error {
	// 获取当前关注状态
	if isFollow, err := GetFollowStatusForUpdate(followerID, celebrityID); err == nil {
		if !isFollow {
			return nil
		}
		// 修改数据库
		if err := global.DB.Model(&model.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
			Update("is_follow", false).Error; err != nil {
			return err
		}
	} else {
		return err
	}
	//更新缓存
	if err := CancelFollowForRedis(followerID, celebrityID); err != nil {
		return err
	}
	return nil
}

// GetFollowIDListByUserID 通过用户 ID 查询关注 ID 列表
func GetFollowIDListByUserID(followerID uint64) ([]uint64, error) {
	// 查询缓存
	celebrityIDList, err := GetFollowIDListByUserIDFromRedis(followerID)
	if err == nil {
		return celebrityIDList, nil
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	// 缓存不存在，查询数据库
	var followList []model.Follow
	if result := global.DB.Model(&model.Follow{}).Select("celebrity_id", "is_follow").
		Where("follower_id = ?", followerID).Find(&followList); result.Error != nil {
		return nil, result.Error
	}
	// 更新缓存
	if err = AddFollowIDListByUserIDToRedis(followerID, followList); err != nil {
		return nil, err
	}
	// 后续操作，返回关注 ID 列表
	celebrityIDList = make([]uint64, 0, len(followList))
	for _, each := range followList {
		celebrityIDList = append(celebrityIDList, each.CelebrityID)
	}
	return celebrityIDList, nil

}

// GetFollowListByUserID 获取用户关注列表
func GetFollowListByUserID(followerID uint64) ([]model.User, error) {
	// 通过用户 ID 查询关注 ID 列表
	celebrityIDList, err := GetFollowIDListByUserID(followerID)
	if err != nil {
		return nil, err
	}
	// 后续处理，返回用户关注列表
	celebrityList, err := GetUserListByUserIDList(celebrityIDList)
	if err != nil {
		return nil, err
	}
	return celebrityList, nil
}

// GetFollowerIDListByUserID 通过用户 ID 查询粉丝 ID 列表
func GetFollowerIDListByUserID(celebrityID uint64) ([]uint64, error) {
	// 查询缓存
	followerIDList, err := GetFollowerIDListByUserIDFromRedis(celebrityID)
	if err == nil {
		return followerIDList, nil
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	// 缓存不存在，查询数据库
	var followerList []model.Follow
	result := global.DB.Model(&model.Follow{}).Where("celebrity_id = ? and is_follow = ?", celebrityID, true).
		Find(&followerList)
	if result.Error != nil {
		return nil, result.Error
	}
	// 更新缓存
	if err = AddFollowerIDListByUserIDToRedis(celebrityID, followerList); err != nil {
		return nil, err
	}
	// 后续操作，返回粉丝 ID 列表
	followerIDList = make([]uint64, 0, len(followerList))
	for _, each := range followerList {
		followerIDList = append(followerIDList, each.FollowerID)
	}
	return followerIDList, nil
}

// GetFollowerListByUserID 获取用户粉丝列表
func GetFollowerListByUserID(celebrityID uint64) ([]model.User, error) {
	// 通过用户 ID 查询粉丝 ID 列表
	followerIDList, err := GetFollowerIDListByUserID(celebrityID)
	if err != nil {
		return nil, err
	}
	// 后续处理，返回用户粉丝列表
	followerList, err := GetUserListByUserIDList(followerIDList)
	if err != nil {
		return nil, err
	}
	return followerList, nil
}

// GetFollowStatusList 返回关注状态列表
func GetFollowStatusList(followerID uint64, celebrityIDList []uint64) ([]bool, error) {
	// 通过用户 ID 查询粉丝 ID 列表
	CelebrityIDList, err := GetFollowIDListByUserID(followerID)
	if err != nil {
		return nil, err
	}
	// 后续处理，返回关注状态列表
	mapCelebrityIDToIsFollow := make(map[uint64]void, len(CelebrityIDList))
	for _, each := range CelebrityIDList {
		mapCelebrityIDToIsFollow[each] = member
	}
	isFollowStatusList := make([]bool, len(celebrityIDList)) // 返回结果
	for i, each := range celebrityIDList {
		if _, ok := mapCelebrityIDToIsFollow[each]; ok {
			isFollowStatusList[i] = true
		}
	}
	return isFollowStatusList, nil
}
