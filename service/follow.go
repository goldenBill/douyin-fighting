package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
)

// AddFollow 关注
func AddFollow(followerID, celebrityID uint64) error {
	return global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		follow := dao.Follow{}
		// 得到结果
		result := tx.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).Limit(1).Find(&follow)
		// 数据库中的条目存在 且 没有关注
		if result.RowsAffected != 0 && !follow.IsFollow {
			// 更新关注状态
			follow.IsFollow = true
			if err := tx.Save(&follow).Error; err != nil {
				return err
			}
		} else if result.RowsAffected == 0 {
			// 在关注表中新增一个条目
			follow.FollowID, _ = global.GVAR_ID_GENERATOR.NextID()
			follow.CelebrityID = celebrityID
			follow.FollowerID = followerID
			follow.IsFollow = true
			if err := tx.Create(&follow).Error; err != nil {
				return err
			}
		} else {
			return nil
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
		follow := dao.Follow{}
		// 得到结果
		result := tx.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).Limit(1).Find(&follow)
		// 数据库中的条目存在 且 有关注
		if result.RowsAffected == 0 || !follow.IsFollow {
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
	followList := make([]dao.Follow, 0, 20)
	celebrityIDList := make([]uint64, 0, 20)
	global.GVAR_DB.Model(&dao.Follow{}).Where("follower_id = ? and is_follow = ?", userID, true).Find(&followList)
	for _, each := range followList {
		celebrityIDList = append(celebrityIDList, each.CelebrityID)
	}
	var celebrityList []dao.User
	err := GetUserListByUserIDs(celebrityIDList, celebrityList)
	if err != nil {
		return nil, err
	}
	return celebrityList, nil
}

// GetFollowerListByUserID 获取用户粉丝列表
func GetFollowerListByUserID(userID uint64) ([]dao.User, error) {
	followList := make([]dao.Follow, 0, 20)
	followerIDList := make([]uint64, 0, 20)
	global.GVAR_DB.Model(&dao.Follow{}).Where("celebrity_id = ? and is_follow = ?", userID, true).Find(&followList)
	for _, each := range followList {
		followerIDList = append(followerIDList, each.FollowerID)
	}
	var followerList []dao.User
	err := GetUserListByUserIDs(followerIDList, followerList)
	if err != nil {
		return nil, err
	}
	return followerList, nil
}

// GetIsFollowStatus 根据 celebrityID 和 followerID 返回关注状态
func GetIsFollowStatus(followerID, celebrityID uint64) bool {
	var follow dao.Follow
	// 得到结果
	global.GVAR_DB.Model(&dao.Follow{}).Where("celebrity_id = ? and follower_id = ?", celebrityID, followerID).Limit(1).Find(&follow)
	return follow.IsFollow
}

// GetIsFollowStatusList 根据 celebrityIDList 和 followerID 返回关注状态
func GetIsFollowStatusList(followerID uint64, celebrityIDList []uint64) ([]bool, error) {
	var uniqueFollows []dao.Follow
	result := global.GVAR_DB.Model(&dao.Follow{}).Where("celebrity_id in ? and follower_id = ?", celebrityIDList, followerID).Find(&uniqueFollows)
	if result.Error != nil {
		err := errors.New("query GetIsFollowStatusList error")
		return nil, err
	}
	mapCelebrityIDToIsFollow := make(map[uint64]bool, len(uniqueFollows))
	for _, follow := range uniqueFollows {
		mapCelebrityIDToIsFollow[follow.CelebrityID] = follow.IsFollow
	}
	isFollowStatusList := make([]bool, len(celebrityIDList))
	for idx, celebrityID := range celebrityIDList {
		isFollowStatusList[idx] = mapCelebrityIDToIsFollow[celebrityID]
	}
	return isFollowStatusList, nil
}
