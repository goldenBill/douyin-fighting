package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"time"
)

// 按要求拉取feed视频
func GetFeedVideos(videos *[]dao.Video, LatestTime int64, MaxNumVideo int) *gorm.DB {
	result := global.GVAR_DB.Debug().Where("created_at < ?", time.UnixMilli(LatestTime)).Order("created_at DESC").Limit(MaxNumVideo).Find(videos)
	return result
}

// 记录接收视频的属性并写入数据库
func PublishVideo(userID uint64, videoID uint64, videoName string, coverName string, title string) error {
	var video dao.Video
	video.VideoID = videoID
	video.Title = title
	video.PlayName = videoName
	video.CoverName = coverName
	//video.FavoriteCount = 0
	//video.CommentCount = 0
	video.AuthorID = userID
	//video.CreatedAt = time.Now().UnixMilli()

	err := global.GVAR_DB.Debug().Create(&video).Error
	return err
}

// 给定用户ID得到其发表过的视频
func GetPublishedVideos(videos *[]dao.Video, userID uint64) error {
	err := global.GVAR_DB.Debug().Where("author_id = ?", userID).Find(videos).Error
	return err
}

// 给定视频ID列表得到对应的视频信息
func GetVideoListByIDs(videos *[]dao.Video, videoIDs []uint64) error {
	err := global.GVAR_DB.Debug().Where("video_id in (?)", videoIDs).Find(videos).Error
	return err
}
