package service

import (
	"fmt"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

// FavoriteAction 用户userID进行点赞视频videoID的操作
func FavoriteAction(userID, videoID uint64) error {
	f := dao.Favorite{}

	// 得到结果
	result := global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).First(&f)

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
	result := global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).First(&f)

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
		fmt.Printf("132456798%#v\n", err)
		return []dao.Video{}, err
	}
	return videoDaoList, nil
}

// GetFavoriteStatus 获取用户userID是否点赞了视频videoID
func GetFavoriteStatus(userID, videoID uint64) bool {
	var f dao.Favorite
	// 得到结果
	global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).First(&f)
	return f.IsFavorite
}
