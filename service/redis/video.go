package redis

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

// GetVideoListByIDs 给定视频ID列表得到对应的视频信息
func GetVideoListByIDs(videos *[]dao.Video, videoIDs []uint64) error {
	err := global.GVAR_DB.Where("video_id in ?", videoIDs).Find(videos).Error
	return err
}
