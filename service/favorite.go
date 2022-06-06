package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
)

func GetFavoriteStatusForUpdate(userID, videoID uint64) (bool, error) {
	favoriteStatus, err := GetFavoriteStatusFromRedis(userID, videoID)
	if err == nil {
		return favoriteStatus, nil
	} else if err.Error() != "not found in cache" {
		return false, err
	}
	//缓存不存在，查询数据库
	var favoriteList []model.Favorite
	if result := global.DB.Select("video_id", "is_favorite").Model(&model.Favorite{}).
		Where("user_id = ?", userID).Find(&favoriteList); result.Error != nil {
		return false, result.Error
	}
	//更新 redis
	if err = AddFavoriteVideoIDListByUserIDToRedis(userID, favoriteList); err != nil {
		return false, err
	}
	return GetFavoriteStatusFromRedis(userID, videoID)
}

func GetFavoriteStatus(userID, videoID uint64) (bool, error) {
	isFavorite, err := GetFavoriteStatusForUpdate(userID, videoID)
	if err == nil || err.Error() == "no tracking information" {
		return isFavorite, nil
	}
	return false, err
}

func AddFavorite(userID, videoID uint64) error {
	if isFavorite, err := GetFavoriteStatusForUpdate(userID, videoID); err == nil {
		if isFavorite {
			return nil
		}
		if err := global.DB.Model(&model.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).
			Update("is_favorite", true).Error; err != nil {
			return err
		}
	} else if err.Error() == "no tracking information" {
		var favorite model.Favorite
		//数据库
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
	//更新缓存
	if err := AddFavoriteForRedis(videoID, userID, video.AuthorID); err != nil {
		return err
	}
	return nil
}

// CancelFavorite 用户userID取消点赞视频videoID
func CancelFavorite(userID, videoID uint64) error {
	if isFavorite, err := GetFavoriteStatusForUpdate(userID, videoID); err == nil {
		if !isFavorite {
			return nil
		}
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
	//更新缓存
	if err := CancelFavoriteForRedis(videoID, userID, video.AuthorID); err != nil {
		return err
	}
	return nil
}

func GetFavoriteVideoIDListByUserID(userID uint64) ([]uint64, error) {
	//查询redis
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserIDFromRedis(userID)
	if err == nil {
		return favoriteVideoIDList, nil
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	//redis没找到，数据库查询
	var favoriteList []model.Favorite
	if result := global.DB.Select("video_id", "is_favorite").Model(&model.Favorite{}).
		Where("user_id = ?", userID).Find(&favoriteList); result.Error != nil {
		return nil, result.Error
	}
	//更新 redis
	if err = AddFavoriteVideoIDListByUserIDToRedis(userID, favoriteList); err != nil {
		return nil, err
	}
	//后续操作
	favoriteVideoIDList = make([]uint64, 0, len(favoriteList))
	for _, each := range favoriteList {
		favoriteVideoIDList = append(favoriteVideoIDList, each.VideoID)
	}
	return favoriteVideoIDList, nil
}

// GetFavoriteListByUserID 获取用户点赞列表
func GetFavoriteListByUserID(userID uint64) ([]model.Video, error) {
	favoriteVideoIDList, err := GetFavoriteVideoIDListByUserID(userID)
	if err != nil {
		return nil, err
	}
	//后续处理
	var videoList []model.Video
	err = GetVideoListByIDsRedis(&videoList, favoriteVideoIDList)
	if err != nil {
		return nil, err
	}
	return videoList, nil
}

// GetFavoriteStatusList 根据 userID 和 videoIDList 返回点赞状态（列表）
func GetFavoriteStatusList(userID uint64, videoIDList []uint64) ([]bool, error) {
	//查询redis
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

func GetFavoriteCountByVideoID(videoID uint64) (int64, error) {
	//查询缓存
	favoriteCount, err := GetFavoriteCountByVideoIDFromRedis(videoID)
	if err == nil {
		return favoriteCount, nil
	} else if err.Error() != "not found in cache" {
		return 0, err
	}
	//缓存没有找到，数据库查询
	err = global.DB.Model(&model.Favorite{}).Where("video_id = ? and is_favorite = ?", videoID, true).
		Count(&favoriteCount).Error
	if err != nil {
		return 0, err
	}
	//更新缓存
	if err := AddFavoriteCountByVideoIDToRedis(videoID, favoriteCount); err != nil {
		return 0, err
	}
	return favoriteCount, nil
}

func GetFavoriteCountListByVideoIDList(videoIDList []uint64) ([]int64, error) {
	//查询redis
	favoriteCountList, notInCache, err := GetFavoriteCountListByVideoIDListFromRedis(videoIDList)
	if err == nil {
		return favoriteCountList, nil
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	//缓存没有找到，数据库查询
	var uniqueVideoList []VideoFavoriteCountAPI
	result := global.DB.Model(&model.Favorite{}).Select("video_id", "COUNT(video_id) as favorite_count").
		Where("video_id in ? and is_favorite = ?", notInCache, true).Group("video_id").Find(&uniqueVideoList)
	if result.Error != nil {
		return nil, result.Error
	}
	//更新缓存
	if err = AddFavoriteCountListByUVideoIDListToCache(uniqueVideoList); err != nil {
		return nil, err
	}
	//后续操作
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

func GetFavoriteCountByUserID(userID uint64) (int64, error) {
	//查询缓存
	favoriteCount, err := GetFavoriteCountByUserIDFromRedis(userID)
	if err == nil {
		return favoriteCount, nil
	} else if err.Error() != "not found in cache" {
		return 0, err
	}
	//缓存没有找到，数据库查询
	err = global.DB.Model(&model.Favorite{}).Where("user_id = ? and is_favorite = ?", userID, true).
		Count(&favoriteCount).Error
	if err != nil {
		return 0, err
	}
	//更新缓存
	if err = AddFavoriteCountByUserIDToRedis(userID, favoriteCount); err != nil {
		return 0, err
	}
	return favoriteCount, nil
}

func GetTotalFavoritedByUserID(userID uint64) (int64, error) {
	//查询缓存
	totalFavorited, err := GetTotalFavoritedByUserIDFromRedis(userID)
	if err == nil {
		return totalFavorited, nil
	} else if err.Error() != "not found in cache" {
		return 0, err
	}
	//缓存不存在
	var publishVideoIDList []uint64
	if err = GetVideoIDListByUserID(userID, &publishVideoIDList); err != nil {
		return 0, err
	}
	favoriteCountList, err := GetFavoriteCountListByVideoIDList(publishVideoIDList)
	if err != nil {
		return 0, err
	}
	totalFavorited = 0
	for _, each := range favoriteCountList {
		totalFavorited += each
	}
	//更新缓存
	if err = AddTotalFavoritedByUserIDToRedis(userID, totalFavorited); err != nil {
		return 0, err
	}
	return totalFavorited, nil
}
