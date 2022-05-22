package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"time"
)

type Video struct {
}

func (*Video) GetVideos(LatestTime time.Time) (videos []dao.Video, rowsAffected int64) {
	rowsAffected = global.GVAR_DB.Debug().Where("created_at < ?", LatestTime).Where("active = ?", true).
		Order("created_at DESC").Limit(global.GVAR_FEED_NUM).Find(&videos).RowsAffected
	return
}

func (*Video) PublishVideo(video *dao.Video) error {
	//生成增长的 VideoID
	video.VideoID, _ = global.GVAR_ID_GENERATOR.NextID()
	err := global.GVAR_DB.Debug().Create(video).Error
	return err
}

func (*Video) SetActive(videoID uint64) error {
	err := global.GVAR_DB.Debug().Model(&dao.Video{}).Where("video_id = ?", videoID).Update("active", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (*Video) GetPublishedVideos(userID uint64) (videos []dao.Video, err error) {
	err = global.GVAR_DB.Debug().Where("user_id = ?", userID).Where("active = ?", true).Find(&videos).Error
	return
}
