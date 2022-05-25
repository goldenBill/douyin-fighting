package service

import (
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

// GetFavoriteStatusList 根据userIDList和，videoIDList 返回点赞状态（列表）
func GetFavoriteStatusList(userIDList, videoIDList []uint64) []bool {
	type pair struct {
		// 记录(用户ID, 视频ID)二元组的结构体
		userID  uint64
		videoID uint64
	}
	isFavoriteList := make([]bool, 0, len(userIDList)) // 返回结果
	hashMap := make(map[pair]bool, len(userIDList))    // 记录重复结果
	for i := 0; i < len(userIDList); i++ {
		userID, videoID := userIDList[i], videoIDList[i]
		isFavorite, ok := hashMap[pair{userID: userID, videoID: videoID}]
		if !ok {
			// 在map中没有，则查询数据库
			var f dao.Favorite
			global.GVAR_DB.Model(&dao.Favorite{}).Where("user_id = ? and video_id = ?", userID, videoID).Limit(1).Find(&f)
			isFavoriteList = append(isFavoriteList, f.IsFavorite)
			hashMap[pair{userID: userID, videoID: videoID}] = f.IsFavorite
		} else {
			// 直接读取 map 数据
			isFavoriteList = append(isFavoriteList, isFavorite)
		}
	}
	return isFavoriteList
}
