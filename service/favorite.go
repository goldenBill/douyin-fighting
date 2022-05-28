package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
)

// FavoriteAction 用户userID进行点赞视频videoID的操作
func FavoriteAction(userID, videoID uint64) error {
	return global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 得到结果
		var favorite dao.Favorite
		result := tx.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).
			Limit(1).Find(&favorite)
		//数据库中的条目存在
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected != 0 {
			if favorite.IsFavorite {
				return nil
			}
			// 更新点赞状态
			favorite.IsFavorite = true
			if err := tx.Save(&favorite).Error; err != nil {
				return err
			}
		}
		// 在点赞表中新增一个条目
		favorite.FavoriteID, _ = global.GVAR_ID_GENERATOR.NextID()
		favorite.VideoID = videoID
		favorite.UserID = userID
		favorite.IsFavorite = true
		if err := tx.Create(&favorite).Error; err != nil {
			// 插入出错，直接返回
			return err
		}
		// 更新 videos 表的 FavoriteCount
		if err := tx.Model(&dao.Video{}).Where("video_id = ?", videoID).
			Update("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
			return err
		}
		// 更新 users 表的 FavoriteCount
		if err := tx.Model(&dao.User{}).Where("user_id = ?", userID).
			Update("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
			return err
		}
		// 查询视频作者
		var video dao.Video
		if result := tx.Where("video_id = ?", videoID).Limit(1).Find(&video); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return errors.New("video 表中 video_id 不存在")
		}
		// 更新 users 表的 TotalFavorited
		if err := tx.Model(&dao.User{}).Where("user_id = ?", video.AuthorID).
			Update("total_favorited", gorm.Expr("total_favorited + 1")).Error; err != nil {
			return err
		}
		//删除 redis 缓存
		if err := DeleteFavoriteFromCache(videoID, userID, video.AuthorID); err != nil {
			return err
		}
		return nil
	})
}

// CancelFavorite 用户userID取消点赞视频videoID
func CancelFavorite(userID, videoID uint64) error {
	return global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 得到结果
		var favorite dao.Favorite
		result := tx.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).
			Limit(1).Find(&favorite)
		//数据库中的条目不存在
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 || !favorite.IsFavorite {
			return nil
		}
		// 更新点赞状态
		favorite.IsFavorite = false
		if err := tx.Save(&favorite).Error; err != nil {
			return err
		}
		// 更新 videos 表的 FavoriteCount
		if err := tx.Model(&dao.Video{}).Where("video_id = ?", videoID).
			Update("favorite_count", gorm.Expr("favorite_count - 1")).Error; err != nil {
			return err
		}
		// 更新 users 表的 FavoriteCount
		if err := tx.Model(&dao.User{}).Where("user_id = ?", userID).
			Update("favorite_count", gorm.Expr("favorite_count - 1")).Error; err != nil {
			return err
		}
		// 查询视频作者
		var video dao.Video
		if result := tx.Where("video_id = ?", videoID).Limit(1).Find(&video); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return errors.New("video 表中 video_id 不存在")
		}
		// 更新 users 表的 TotalFavorited
		if err := tx.Model(&dao.User{}).Where("user_id = ?", video.AuthorID).
			Update("total_favorited", gorm.Expr("total_favorited - 1")).Error; err != nil {
			return err
		}
		//删除 redis 缓存
		if err := DeleteFavoriteFromCache(videoID, userID, video.AuthorID); err != nil {
			return err
		}
		return nil
	})
}

func GetFavoriteVideoIDListByUserID(userID uint64) ([]uint64, error) {
	//查询redis
	favoriteVideoIDList, err := GetFavoriteListByUserIDFromCache(userID)
	if err == nil {
	} else if err.Error() != "Not found in cache" {
		return nil, err
	} else {
		//redis没找到，数据库查询
		var favoriteList []dao.Favorite
		result := global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and is_favorite = ?", userID, true).
			Find(&favoriteList)
		if result.Error != nil {
			return nil, result.Error
		}
		favoriteVideoIDList = make([]uint64, 0, len(favoriteList))
		for _, each := range favoriteList {
			favoriteVideoIDList = append(favoriteVideoIDList, each.VideoID)
		}
		//更新 redis
		if err := AddFavoriteListByUserIDInCache(userID, favoriteVideoIDList); err != nil {
			return nil, err
		}
	}
	return favoriteVideoIDList, nil
}

// GetFavoriteListByUserID 获取用户点赞列表
func GetFavoriteListByUserID(userID uint64) ([]dao.Video, error) {
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserID(userID)
	if err != nil {
		return nil, err
	}
	//后续处理
	var videoDaoList []dao.Video
	err = GetVideoListByIDs(&videoDaoList, favoriteVideoIDList)
	if err != nil {
		return nil, err
	}
	return videoDaoList, nil
}

// GetFavoriteStatusList 根据userID和，videoIDList 返回点赞状态（列表）
func GetFavoriteStatusList(userID uint64, videoIDList []uint64) ([]bool, error) {
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserID(userID)
	if err != nil {
		return nil, err
	}
	//后续处理
	mapVideoIDToFavorite := make(map[uint64]void, len(favoriteVideoIDList))
	for _, each := range favoriteVideoIDList {
		mapVideoIDToFavorite[each] = member
	}
	isFavoriteList := make([]bool, len(videoIDList)) // 返回结果
	for i, each := range videoIDList {
		if _, ok := mapVideoIDToFavorite[each]; ok {
			isFavoriteList[i] = true
		}
	}
	return isFavoriteList, nil
}

//// GetFavoriteStatus 获取用户userID是否点赞了视频videoID
//func GetFavoriteStatus(userID, videoID uint64) bool {
//	var favorite dao.Favorite
//	// 得到结果
//	global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).
//		Limit(1).Find(&favorite)
//	return favorite.IsFavorite
//}
