package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"time"
)

func GetVideos(videos *[]dao.Video, LatestTime int64, MaxNumVideo int) *gorm.DB {
	result := global.GVAR_DB.Debug().Where("created_at < ?", LatestTime).Order("created_at DESC").Limit(MaxNumVideo).Find(videos)
	return result
}

func GetAuthor(author *dao.UserForFeed, userID uint64) error {
	err := global.GVAR_DB.Debug().Where("user_id = ?", userID).Take(author).Error
	return err
}

func PublishVideo(userID uint64, videoID uint64, videoName string, coverName string, title string) error {
	var video dao.Video
	video.VideoID = videoID
	video.Title = title
	video.PlayName = videoName
	video.CoverName = coverName
	//video.FavoriteCount = 0
	//video.CommentCount = 0
	video.UserID = userID
	video.CreatedAt = time.Now().UnixMilli()

	err := global.GVAR_DB.Debug().Create(&video).Error
	return err
}

func GetPublishedVideos(videos *[]dao.Video, userID uint64) error {
	err := global.GVAR_DB.Debug().Where("user_id = ?", userID).Find(videos).Error
	return err
}

func GetVideoListByIDs(videos *[]dao.Video, videoIDs []uint64) error {
	err := global.GVAR_DB.Debug().Where("video_id in (?)", videoIDs).Find(videos).Error
	return err
}
