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
			result = global.GVAR_DB.Save(&f)
			return result.Error
		}
	} else {
		// 在点赞表中新增一个条目
		f.VideoID = videoID
		f.UserID = userID
		f.IsFavorite = true
		result = global.GVAR_DB.Create(&f)
		return result.Error
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
			result = global.GVAR_DB.Save(&f)
			return result.Error
		}
	}
	// 不存在直接忽略

	return nil
}

// GetFavoriteListByUserID 获取用户点赞列表
func GetFavoriteListByUserID(userID uint64) ([]dao.Video, error) {
	favoriteList := make([]dao.Favorite, 0, 20)
	videoIDList := make([]uint64, 0, 20)
	global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and is_favorite = ?", userID, true).Find(&favoriteList)
	for _, each := range favoriteList {
		videoIDList = append(videoIDList, each.VideoID)
	}

	videoDaoList, err := GetVideoListByVideoIDs(videoIDList)
	if err != nil {
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

// GetFavoriteCount 获取视频videoID的点赞数
func GetFavoriteCount(videoID uint64) uint64 {
	var count int64
	global.GVAR_DB.Model(&dao.Favorite{}).Where("video_id = ? and is_favorite = ?", videoID, true).Count(&count)
	return uint64(count)
}

func GetVideoListByVideoIDs(videoIDList []uint64) ([]dao.Video, error) {
	var videoDaoList []dao.Video
	for _, videoID := range videoIDList {
		var videoDao dao.Video
		sqlStr := "select * from video where video_id = ?"
		if err := global.GVAR_SQLX_DB.Get(&videoDao, sqlStr, videoID); err != nil {
			return []dao.Video{}, err
		}
		videoDaoList = append(videoDaoList, videoDao)
	}

	return videoDaoList, nil
}

func GetFollowCount(userID uint64) uint64 {
	var followCount int64
	sqlStr := "select count(*) from follow where follower_id = ?"
	if err := global.GVAR_SQLX_DB.Get(&followCount, sqlStr, userID); err != nil {
		fmt.Println("exec failed, ", err)
		return 0
	}
	return uint64(followCount)
}

func GetFollowerCount(userID uint64) uint64 {
	var followCount int64
	sqlStr := "select count(*) from follow where user_id = ?"
	if err := global.GVAR_SQLX_DB.Get(&followCount, sqlStr, userID); err != nil {
		fmt.Println("exec failed, ", err)
		return 0
	}
	return uint64(followCount)
}
