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
		// 返回 nil 提交事务
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
		// 返回 nil 提交事务
		return nil
	})
}

// GetFollowListByUserID 获取用户关注列表
func GetFollowListByUserID(userID uint64) ([]dao.User, error) {
	var followList []dao.Follow
	result := global.GVAR_DB.Model(&dao.Follow{}).Where("follower_id = ? and is_follow = ?", userID, true).
		Find(&followList)
	if result.Error != nil {
		return nil, result.Error
	}
	celebrityIDList := make([]uint64, 0, len(followList))
	for _, each := range followList {
		celebrityIDList = append(celebrityIDList, each.CelebrityID)
	}
	celebrityList, err := GetUserListByUserIDList(celebrityIDList)
	if err != nil {
		return nil, err
	}
	return celebrityList, nil
}

// GetFollowerListByUserID 获取用户粉丝列表
func GetFollowerListByUserID(userID uint64) ([]dao.User, error) {
	var followList []dao.Follow
	result := global.GVAR_DB.Model(&dao.Follow{}).Where("celebrity_id = ? and is_follow = ?", userID, true).
		Find(&followList)
	if result.Error != nil {
		return nil, result.Error
	}
	followerIDList := make([]uint64, 0, len(followList))
	for _, each := range followList {
		followerIDList = append(followerIDList, each.FollowerID)
	}
	followerList, err := GetUserListByUserIDList(followerIDList)
	if err != nil {
		return nil, err
	}
	return followerList, nil
}

// GetIsFollowStatus 根据 celebrityID 和 followerID 返回关注状态
func GetIsFollowStatus(followerID, celebrityID uint64) (bool, error) {
	var follow dao.Follow
	// 得到结果
	result := global.GVAR_DB.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).
		Limit(1).Find(&follow)
	if result.Error != nil {
		return false, result.Error
	}
	return follow.IsFollow, nil
}

// GetIsFollowStatusList 根据 celebrityIDList 和 followerID 返回关注状态
func GetIsFollowStatusList(followerID uint64, celebrityIDList []uint64) ([]bool, error) {
	var uniqueFollows []dao.Follow
	result := global.GVAR_DB.Model(&dao.Follow{}).
		Where("celebrity_id in ? and follower_id = ?", celebrityIDList, followerID).Find(&uniqueFollows)
	if result.Error != nil {
		return nil, result.Error
	}
	//针对查询建立映射关系
	mapCelebrityIDToIsFollow := make(map[uint64]bool, len(uniqueFollows))
	for _, follow := range uniqueFollows {
		mapCelebrityIDToIsFollow[follow.CelebrityID] = follow.IsFollow
	}
	//构造返回值
	isFollowStatusList := make([]bool, len(celebrityIDList))
	for idx, celebrityID := range celebrityIDList {
		isFollowStatusList[idx] = mapCelebrityIDToIsFollow[celebrityID]
	}
	return isFollowStatusList, nil
}
