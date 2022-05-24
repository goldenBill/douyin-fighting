package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"time"
)

// GetFeedVideos 按要求拉取feed视频
func GetFeedVideos(videos *[]dao.Video, LatestTime int64, MaxNumVideo int) *gorm.DB {
	result := global.GVAR_DB.Debug().Where("created_at < ?", time.UnixMilli(LatestTime)).Order("created_at DESC").Limit(MaxNumVideo).Find(videos)
	return result
}

// PublishVideo 记录接收视频的属性并写入数据库
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

// GetPublishedVideos 给定用户ID得到其发表过的视频
func GetPublishedVideos(videos *[]dao.Video, userID uint64) error {
	err := global.GVAR_DB.Debug().Where("author_id = ?", userID).Find(videos).Error
	return err
}

// GetVideoListByIDs 给定视频ID列表得到对应的视频信息
func GetVideoListByIDs(videos *[]dao.Video, videoIDs []uint64) error {
	err := global.GVAR_DB.Debug().Where("video_id in ?", videoIDs).Find(videos).Error
	return err
}

// FavoriteCountPlus 点赞数加1
func FavoriteCountPlus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
	return err
}

// FavoriteCountMinus 点赞数减1
func FavoriteCountMinus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count - 1")).Error
	return err
}

// CommentCountPlus 评论数加1
func CommentCountPlus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count + 1")).Error
	return err
}

// CommentCountMinus 评论数减1
func CommentCountMinus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count - 1")).Error
	return err
}
