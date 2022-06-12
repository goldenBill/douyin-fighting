package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
)

// GetFavoriteStatusForUpdate 获取点赞状态，此处是针对 AddFavorite 和 CancelFavorite
func GetFavoriteStatusForUpdate(userID, videoID uint64) (bool, error) {
	// 查询缓存
	favoriteStatus, err := GetFavoriteStatusFromRedis(userID, videoID)
	if err == nil {
		return favoriteStatus, nil
	} else if err.Error() != "not found in cache" {
		return false, err
	}
	// 缓存不存在，查询数据库
	var favoriteList []model.Favorite
	if result := global.DB.Select("video_id", "is_favorite").Model(&model.Favorite{}).
		Where("user_id = ?", userID).Find(&favoriteList); result.Error != nil {
		return false, result.Error
	}
	// 更新缓存
	if err = AddFavoriteVideoIDListByUserIDToRedis(userID, favoriteList); err != nil {
		return false, err
	}
	return GetFavoriteStatusFromRedis(userID, videoID)
}

// AddFavorite 点赞
func AddFavorite(userID, videoID uint64) error {
	// 获取当前点赞状态
	if isFavorite, err := GetFavoriteStatusForUpdate(userID, videoID); err == nil {
		if isFavorite {
			return nil
		}
		// 数据库有记录，修改数据库
		if err := global.DB.Model(&model.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).
			Update("is_favorite", true).Error; err != nil {
			return err
		}
	} else if err.Error() == "no tracking information" {
		var favorite model.Favorite
		// 数据库没有记录，写入数据库
		favorite.FavoriteID, _ = global.ID_GENERATOR.NextID()
		favorite.VideoID = videoID
		favorite.UserID = userID
		favorite.IsFavorite = true
		if err := global.DB.Create(&favorite).Error; err != nil {
			// 插入出错，直接返回
			return err
		}
	} else {
		return err
	}
	// 查询视频作者
	var video model.Video
	if result := global.DB.Select("author_id").Where("video_id = ?", videoID).Limit(1).
		Find(&video); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return errors.New("video 表中 video_id 不存在")
	}
	// 更新缓存
	if err := AddFavoriteForRedis(videoID, userID, video.AuthorID); err != nil {
		return err
	}
	return nil
}

// CancelFavorite 取消点赞
func CancelFavorite(userID, videoID uint64) error {
	// 获取当前点赞状态
	if isFavorite, err := GetFavoriteStatusForUpdate(userID, videoID); err == nil {
		if !isFavorite {
			return nil
		}
		// 修改数据库
		if err := global.DB.Model(&model.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).
			Update("is_favorite", false).Error; err != nil {
			return err
		}
	} else {
		return err
	}
	// 查询视频作者
	var video model.Video
	if result := global.DB.Select("author_id").Where("video_id = ?", videoID).Limit(1).
		Find(&video); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return errors.New("video 表中 video_id 不存在")
	}
	// 更新缓存
	if err := CancelFavoriteForRedis(videoID, userID, video.AuthorID); err != nil {
		return err
	}
	return nil
}

// GetFavoriteVideoIDListByUserID 通过用户 ID 查询点赞视频 ID 列表
func GetFavoriteVideoIDListByUserID(userID uint64) ([]uint64, error) {
	// 查询缓存
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserIDFromRedis(userID)
	if err == nil {
		return favoriteVideoIDList, nil
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	// 缓存不存在，查询数据库
	var favoriteList []model.Favorite
	if result := global.DB.Select("video_id", "is_favorite").Model(&model.Favorite{}).
		Where("user_id = ?", userID).Find(&favoriteList); result.Error != nil {
		return nil, result.Error
	}
	// 更新缓存
	if err = AddFavoriteVideoIDListByUserIDToRedis(userID, favoriteList); err != nil {
		return nil, err
	}
	// 后续操作，返回点赞视频 ID 列表
	favoriteVideoIDList = make([]uint64, 0, len(favoriteList))
	for _, each := range favoriteList {
		favoriteVideoIDList = append(favoriteVideoIDList, each.VideoID)
	}
	return favoriteVideoIDList, nil
}

// GetFavoriteListByUserID 获取用户点赞视频列表
func GetFavoriteListByUserID(userID uint64) ([]model.Video, error) {
	// 通过用户 ID 查询点赞视频 ID 列表
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserID(userID)
	if err != nil {
		return nil, err
	}
	// 后续处理，返回点赞视频列表
	var videoList []model.Video
	err = GetVideoListByIDsRedis(&videoList, favoriteVideoIDList)
	if err != nil {
		return nil, err
	}
	return videoList, nil
}

// GetFavoriteStatusList 根据 userID 和 videoIDList 返回点赞状态列表
func GetFavoriteStatusList(userID uint64, videoIDList []uint64) ([]bool, error) {
	// 通过用户 ID 查询点赞视频 ID 列表
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserID(userID)
	if err != nil {
		return nil, err
	}
	// 后续处理，返回点赞状态列表
	mapVideoIDToFavorite := make(map[uint64]void, len(favoriteVideoIDList))
	for _, each := range favoriteVideoIDList {
		mapVideoIDToFavorite[each] = member
	}
	isFavoriteList := make([]bool, len(videoIDList))
	for i, each := range videoIDList {
		if _, ok := mapVideoIDToFavorite[each]; ok {
			isFavoriteList[i] = true
		}
	}
	return isFavoriteList, nil
}

// GetFavoriteCountListByVideoIDList 根据视频 ID 列表返回点赞数量列表
func GetFavoriteCountListByVideoIDList(videoIDList []uint64) ([]int64, error) {
	// 查询缓存
	favoriteCountList, notInCache, err := GetFavoriteCountListByVideoIDListFromRedis(videoIDList)
	if err == nil {
		return favoriteCountList, nil
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	// 缓存没有找到，数据库查询
	var uniqueVideoList []VideoFavoriteCountAPI
	result := global.DB.Model(&model.Favorite{}).Select("video_id", "COUNT(video_id) as favorite_count").
		Where("video_id in ? and is_favorite = ?", notInCache, true).Group("video_id").Find(&uniqueVideoList)
	if result.Error != nil {
		return nil, result.Error
	}
	// 更新缓存
	if err = AddFavoriteCountListByUVideoIDListToCache(uniqueVideoList); err != nil {
		return nil, err
	}
	// 后续操作，返回点赞数量列表
	// 针对查询结果建立映射关系
	mapVideoIDToFavoriteCount := make(map[uint64]int64, len(uniqueVideoList))
	for _, each := range uniqueVideoList {
		mapVideoIDToFavoriteCount[each.VideoID] = each.FavoriteCount
	}
	scanner := 0
	for idx, each := range favoriteCountList {
		if each == -1 {
			favoriteCountList[idx] = mapVideoIDToFavoriteCount[notInCache[scanner]]
			scanner++
		}
	}
	return favoriteCountList, nil
}
