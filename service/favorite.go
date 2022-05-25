package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

// FavoriteAction 用户userID进行点赞视频videoID的操作
func FavoriteAction(userID, videoID uint64) error {
	f := dao.Favorite{}

	// 得到结果
	result := global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).Limit(1).Find(&f)

	if result.RowsAffected != 0 {
		// 数据库中的条目存在
		if f.IsFavorite == false {
			// 更新点赞状态
			f.IsFavorite = true
			err := global.GVAR_DB.Save(&f).Error
			if err != nil {
				// 更新点赞状态出错，直接返回
				return err
			}
			return FavoriteCountPlus(videoID) // 点赞数+1
		}
	} else {
		// 在点赞表中新增一个条目
		f.FavoriteID, _ = global.GVAR_ID_GENERATOR.NextID()
		f.VideoID = videoID
		f.UserID = userID
		f.IsFavorite = true
		err := global.GVAR_DB.Create(&f).Error
		if err != nil {
			// 插入出错，直接返回
			return err
		}
		return FavoriteCountPlus(videoID) // 点赞数+1
	}

	return nil
}

// CancelFavorite 用户userID取消点赞视频videoID
func CancelFavorite(userID, videoID uint64) error {
	var f dao.Favorite
	// 得到结果
	result := global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).Limit(1).Find(&f)

	if result.RowsAffected != 0 {
		// 数据库中的条目存在
		if f.IsFavorite == true {
			// 更新点赞状态，取消点赞
			f.IsFavorite = false
			err := global.GVAR_DB.Save(&f).Error
			if err != nil {
				// 更新出错，直接返回
				return err
			}
			return FavoriteCountMinus(videoID) // 点赞数-1
		}
	}
	// 不存在直接忽略

	return nil
}

// GetFavoriteListByUserID 获取用户点赞列表
func GetFavoriteListByUserID(userID uint64) ([]dao.Video, error) {
	favoriteList := make([]dao.Favorite, 0, 20)
	videoDaoList := make([]dao.Video, 0, 20)
	videoIDList := make([]uint64, 0, 20)
	global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and is_favorite = ?", userID, true).Find(&favoriteList)
	for _, each := range favoriteList {
		videoIDList = append(videoIDList, each.VideoID)
	}

	err := GetVideoListByIDs(&videoDaoList, videoIDList)
	if err != nil {
		return []dao.Video{}, err
	}
	return videoDaoList, nil
}

// GetFavoriteStatus 获取用户userID是否点赞了视频videoID
func GetFavoriteStatus(userID, videoID uint64) bool {
	var f dao.Favorite
	// 得到结果
	global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).Limit(1).Find(&f)
	return f.IsFavorite
}

// GetFavoriteStatusList 根据userID和，videoIDList 返回点赞状态（列表）
func GetFavoriteStatusList(userID uint64, videoIDList []uint64) ([]bool, error) {
	var f []dao.Favorite
	isFavoriteList := make([]bool, len(videoIDList)) // 返回结果
	result := global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id in (?)", userID, videoIDList).Limit(1).Find(&f)
	if result.Error != nil {
		err := errors.New("query GetFavoriteStatusList error")
		return nil, err
	}
	mapVideoIDToFavorite := make(map[uint64]dao.Favorite)
	for _, favorite := range f {
		mapVideoIDToFavorite[favorite.VideoID] = favorite
	}
	for i, videoID := range videoIDList {
		isFavoriteList[i] = mapVideoIDToFavorite[videoID].IsFavorite
	}
	return isFavoriteList, nil
}
