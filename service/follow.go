package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
)

// AddFollow 关注
func AddFollow(followerID, celebrityID uint64) error {
	return global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 得到结果
		var follow dao.Follow
		result := tx.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
			Limit(1).Find(&follow)
		// 数据库中的条目存在
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected != 0 {
			if follow.IsFollow {
				return nil
			}
			// 更新关注状态
			follow.IsFollow = true
			if err := tx.Save(&follow).Error; err != nil {
				return err
			}
		}
		// 在关注表中新增一个条目
		follow.FollowID, _ = global.GVAR_ID_GENERATOR.NextID()
		follow.CelebrityID = celebrityID
		follow.FollowerID = followerID
		follow.IsFollow = true
		if err := tx.Create(&follow).Error; err != nil {
			return err
		}
		// 更新博主粉丝数
		if err := tx.Model(&dao.User{}).Where("user_id = ?", celebrityID).
			Update("follower_count", gorm.Expr("follower_count + 1")).Error; err != nil {
			return err
		}
		// 更新用户关注数
		if err := tx.Model(&dao.User{}).Where("user_id = ?", followerID).
			Update("follow_count", gorm.Expr("follow_count + 1")).Error; err != nil {
			return err
		}
		//删除 redis 缓存
		if err := DeleteFollowFromCache(followerID, celebrityID); err != nil {
			return err
		}
		return nil
	})
}

// CancelFollow 取消关注
func CancelFollow(followerID, celebrityID uint64) error {
	return global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 得到结果
		var follow dao.Follow
		result := tx.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
			Limit(1).Find(&follow)
		// 数据库中的条目存在 且 有关注
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 || !follow.IsFollow {
			return nil
		}
		// 更新关注状态
		follow.IsFollow = false
		if err := tx.Save(&follow).Error; err != nil {
			return err
		}
		// 更新博主粉丝数
		if err := tx.Model(&dao.User{}).Where("user_id = ?", celebrityID).
			Update("follower_count", gorm.Expr("follower_count - 1")).Error; err != nil {
			return err
		}
		// 更新用户关注数
		if err := tx.Model(&dao.User{}).Where("user_id = ?", followerID).
			Update("follow_count", gorm.Expr("follow_count - 1")).Error; err != nil {
			return err
		}
		//删除 redis 缓存
		if err := DeleteFollowFromCache(followerID, celebrityID); err != nil {
			return err
		}
		return nil
	})
}

func GetFollowIDListByUserID(followerID uint64) ([]uint64, error) {
	//查询redis
	celebrityIDList, err := GetFollowIDListByUserIDFromCache(followerID)
	if err == nil {
	} else if err.Error() != "Not found in cache" {
		return nil, err
	} else {
		//redis没找到，数据库查询
		var followList []dao.Follow
		result := global.GVAR_DB.Model(&dao.Follow{}).Where("follower_id = ? and is_follow = ?", followerID, true).
			Find(&followList)
		if result.Error != nil {
			return nil, result.Error
		}
		celebrityIDList = make([]uint64, 0, len(followList))
		for _, each := range followList {
			celebrityIDList = append(celebrityIDList, each.CelebrityID)
		}
		//更新 redis
		if err := AddFollowIDListByUserIDInCache(followerID, celebrityIDList); err != nil {
			return nil, err
		}
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
	followerIDList, err := GetFollowerIDListByUserIDFromCache(celebrityID)
	if err == nil {
	} else if err.Error() != "Not found in cache" {
		return nil, err
	} else {
		//redis没找到，数据库查询
		var followerList []dao.Follow
		result := global.GVAR_DB.Model(&dao.Follow{}).Where("celebrity_id = ? and is_follow = ?", celebrityID, true).
			Find(&followerList)
		if result.Error != nil {
			return nil, result.Error
		}
		followerIDList = make([]uint64, 0, len(followerList))
		for _, each := range followerList {
			followerIDList = append(followerIDList, each.FollowerID)
		}
		//更新 redis
		if err := AddFollowerIDListByUserIDInCache(celebrityID, followerIDList); err != nil {
			return nil, err
		}
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

// GetIsFollowStatus 根据 celebrityID 和 followerID 返回关注状态
func GetIsFollowStatus(followerID, celebrityID uint64) (bool, error) {
	//查询redis
	allCelebrityIDList, err := GetFollowIDListByUserID(followerID)
	if err != nil {
		return false, err
	}
	//后续处理
	for _, each := range allCelebrityIDList {
		if each == celebrityID {
			return true, nil
		}
	}
	return false, nil
}

// GetIsFollowStatusList 根据 celebrityIDList 和 followerID 返回关注状态
func GetIsFollowStatusList(followerID uint64, celebrityIDList []uint64) ([]bool, error) {
	//查询redis
	allCelebrityIDList, err := GetFollowIDListByUserID(followerID)
	if err != nil {
		return nil, err
	}
	//后续处理
	mapCelebrityIDToIsFollow := make(map[uint64]void, len(allCelebrityIDList))
	for _, each := range allCelebrityIDList {
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
