package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

func GetFollowStatusForUpdate(followerID, celebrityID uint64) (bool, error) {
	followStatus, err := GetFollowStatusFromRedis(followerID, celebrityID)
	if err == nil {
		return followStatus, nil
	} else if err.Error() != "Not found in cache" {
		return false, err
	}
	//缓存不存在，查询数据库
	var followList []dao.Follow
	if result := global.DB.Select("celebrity_id", "is_follow").Model(&dao.Follow{}).
		Where("follower_id = ?", followerID).Find(&followList); result.Error != nil {
		return false, result.Error
	}
	//更新 redis
	if err = AddFollowIDListByUserIDToRedis(followerID, followList); err != nil {
		return false, err
	}
	return GetFollowStatusFromRedis(followerID, celebrityID)
}

func GetFollowStatus(followerID, celebrityID uint64) (bool, error) {
	followStatus, err := GetFollowStatusForUpdate(followerID, celebrityID)
	if err == nil || err.Error() == "No tracking information" {
		return followStatus, nil
	}
	return false, err
}

func AddFollow(followerID, celebrityID uint64) error {
	if isFollow, err := GetFollowStatusForUpdate(followerID, celebrityID); err == nil {
		if isFollow {
			return nil
		}
		if err := global.DB.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
			Update("is_follow", true).Error; err != nil {
			return err
		}
	} else if err.Error() == "No tracking information" {
		var follow dao.Follow
		// 在关注表中新增一个条目
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
	//更新 redis 缓存
	if err := AddFollowForRedis(followerID, celebrityID); err != nil {
		return err
	}
	return nil
}

// CancelFollow 取消关注
func CancelFollow(followerID, celebrityID uint64) error {
	if isFollow, err := GetFollowStatusForUpdate(followerID, celebrityID); err == nil {
		if !isFollow {
			return nil
		}
		if err := global.DB.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
			Update("is_follow", false).Error; err != nil {
			return err
		}
	} else {
		return err
	}
	//更新 redis 缓存
	if err := CancelFollowForRedis(followerID, celebrityID); err != nil {
		return err
	}
	return nil
}

func GetFollowIDListByUserID(followerID uint64) ([]uint64, error) {
	//查询redis
	celebrityIDList, err := GetFollowIDListByUserIDFromRedis(followerID)
	if err == nil {
		return celebrityIDList, nil
	} else if err.Error() != "Not found in cache" {
		return nil, err
	}
	//redis没找到，数据库查询
	var followList []dao.Follow
	if result := global.DB.Model(&dao.Follow{}).Select("celebrity_id", "is_follow").
		Where("follower_id = ?", followerID).Find(&followList); result.Error != nil {
		return nil, result.Error
	}
	//更新 redis
	if err = AddFollowIDListByUserIDToRedis(followerID, followList); err != nil {
		return nil, err
	}
	//后续操作
	celebrityIDList = make([]uint64, 0, len(followList))
	for _, each := range followList {
		celebrityIDList = append(celebrityIDList, each.CelebrityID)
	}
	return celebrityIDList, nil

}

// GetFollowListByUserID 获取用户关注列表
func GetFollowListByUserID(followerID uint64) ([]dao.User, error) {
	celebrityIDList, err := GetFollowIDListByUserID(followerID)
	if err != nil {
		return nil, err
	}
	celebrityList, err := GetUserListByUserIDList(celebrityIDList)
	if err != nil {
		return nil, err
	}
	return celebrityList, nil
}

func GetFollowerIDListByUserID(celebrityID uint64) ([]uint64, error) {
	//查询redis
	followerIDList, err := GetFollowerIDListByUserIDFromRedis(celebrityID)
	if err == nil {
		return followerIDList, nil
	} else if err.Error() != "Not found in cache" {
		return nil, err
	}
	//redis没找到，数据库查询
	var followerList []dao.Follow
	result := global.DB.Model(&dao.Follow{}).Where("celebrity_id = ? and is_follow = ?", celebrityID, true).
		Find(&followerList)
	if result.Error != nil {
		return nil, result.Error
	}
	//更新 redis
	if err = AddFollowerIDListByUserIDToRedis(celebrityID, followerList); err != nil {
		return nil, err
	}
	//后续操作
	followerIDList = make([]uint64, 0, len(followerList))
	for _, each := range followerList {
		followerIDList = append(followerIDList, each.FollowerID)
	}
	return followerIDList, nil
}

// GetFollowerListByUserID 获取用户粉丝列表
func GetFollowerListByUserID(celebrityID uint64) ([]dao.User, error) {
	followerIDList, err := GetFollowerIDListByUserID(celebrityID)
	if err != nil {
		return nil, err
	}
	followerList, err := GetUserListByUserIDList(followerIDList)
	if err != nil {
		return nil, err
	}
	return followerList, nil
}

// GetFollowStatusList 根据 celebrityIDList 和 followerID 返回关注状态
func GetFollowStatusList(followerID uint64, celebrityIDList []uint64) ([]bool, error) {
	//查询redis
	CelebrityIDList, err := GetFollowIDListByUserID(followerID)
	if err != nil {
		return nil, err
	}
	//后续处理
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

func GetFollowCountByUserID(userID uint64) (int64, error) {
	//查询缓存
	followCount, err := GetFollowCountByUserIDFromRedis(userID)
	if err == nil {
		return followCount, nil
	} else if err.Error() != "Not found in cache" {
		return 0, err
	}
	//缓存没有找到，数据库查询
	err = global.DB.Model(&dao.Follow{}).Where("follower_id = ? and is_follow = ?", userID, true).
		Count(&followCount).Error
	if err != nil {
		return 0, err
	}
	//更新缓存
	if err := AddFollowCountByUserIDToRedis(userID, followCount); err != nil {
		return 0, err
	}
	return followCount, nil
}

func GetFollowerCountByUserID(userID uint64) (int64, error) {
	//查询缓存
	followerCount, err := GetFollowerCountByUserIDFromRedis(userID)
	if err == nil {
		return followerCount, nil
	} else if err.Error() != "Not found in cache" {
		return 0, err
	}
	//缓存没有找到，数据库查询
	err = global.DB.Model(&dao.Follow{}).Where("celebrity_id = ? and is_follow = ?", userID, true).
		Count(&followerCount).Error
	if err != nil {
		return 0, err
	}
	//更新缓存
	if err := AddFollowerCountByUserIDToRedis(userID, followerCount); err != nil {
		return 0, err
	}
	return followerCount, nil
}
