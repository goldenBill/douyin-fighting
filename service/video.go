package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
	"time"
)

func GetVideos(videos *[]dao.Video, LatestTime int64, MaxNumVideo int) int64 {
	rowsAffected := global.GVAR_DB.Debug().Where("created_at < ? AND active = ?", LatestTime, true).Order("created_at DESC").Limit(MaxNumVideo).Find(videos).RowsAffected
	return rowsAffected
}

func GetAuthor(author *dao.UserForFeed, userID uint64) error {
	err := global.GVAR_DB.Debug().Where("user_id = ?", userID).Take(author).Error
	return err
}

func PublishVideo(userID uint64, title string) (string, string, uint64, error) {
	var video dao.Video
	video.VideoID, _ = global.GVAR_ID_GENERATOR.NextID()
	video.Title = title
	VideoIDStr := strconv.FormatUint(video.VideoID, 10)
	video.PlayUrl = VideoIDStr + ".mp4"
	video.CoverUrl = VideoIDStr + ".jpg"
	//video.FavoriteCount = 0
	//video.CommentCount = 0
	video.UserID = userID
	video.CreatedAt = time.Now().UnixMilli()
	video.Active = false

	err := global.GVAR_DB.Debug().Create(&video).Error
	return video.PlayUrl, video.CoverUrl, video.VideoID, err
}

func SetActive(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("active", true).Error
	return err
}

func GetPublishedVideos(videos *[]dao.Video, userID uint64) error {
	err := global.GVAR_DB.Debug().Where("user_id = ?", userID).Where("active = ?", true).Find(videos).Error
	return err
}

func GetVideoByID(video *dao.Video, videoID uint64) error {
	err := global.GVAR_DB.Debug().Where("video_id = ?", videoID).Take(video).Error
	return err
}
