package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"time"
)

// 按要求拉取feed视频和其作者
func GetFeedVideosAndAuthors(videos *[]dao.Video, authors *[]dao.User, LatestTime int64, MaxNumVideo int) (int64, error) {
	result := global.GVAR_DB.Debug().Where("created_at < ?", time.UnixMilli(LatestTime)).Order("created_at DESC").Limit(MaxNumVideo).Find(videos)
	if result.Error != nil {
		// 访问数据库出错
		return 0, result.Error
	} else if result.RowsAffected == 0 {
		// 没有满足条件的视频
		return 0, nil
	} else {
		// 成功
		authorIDList := make([]uint64, result.RowsAffected)
		for i, video := range *videos {
			authorIDList[i] = video.AuthorID
		}

		err := GetUserListByUserIDs(authorIDList, authors)
		if err != nil {
			return 0, err
		} else {
			return result.RowsAffected, nil
		}
	}
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

// 按要求拉取feed视频和其作者
func GetPublishedVideosAndAuthors(videos *[]dao.Video, authors *[]dao.User, userID uint64) (int64, error) {
	result := global.GVAR_DB.Debug().Where("author_id = ?", userID).Find(videos)
	if result.Error != nil {
		// 访问数据库出错
		return 0, result.Error
	} else if result.RowsAffected == 0 {
		// 没有满足条件的视频
		return 0, nil
	} else {
		// 成功
		authorIDList := make([]uint64, result.RowsAffected)
		for i, video := range *videos {
			authorIDList[i] = video.AuthorID
		}
		err := GetUserListByUserIDs(authorIDList, authors)
		if err != nil {
			return 0, err
		} else {
			return result.RowsAffected, nil
		}
	}
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

// 点赞数加1
func FavoriteCountPlus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + 1")).Error
	return err
}

// 点赞数减1
func FavoriteCountMinus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count - 1")).Error
	return err
}

// 评论数加1
func CommentCountPlus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count + 1")).Error
	return err
}

// 评论数减1
func CommentCountMinus(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count - 1")).Error
	return err
}
