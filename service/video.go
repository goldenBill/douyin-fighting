package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

type Video struct {
}

func (*Video) AddNewVideo(video *dao.Video) error {
	sqlStr := "insert into video (video_name, video_location, cover_location, uploader_id, title) values (?, ?, ?, ?, ?)"
	_, err := global.GVAR_SQLX_DB.Exec(sqlStr, video.VideoName, video.VideoLocation, video.CoverLocation, video.UploaderId, video.Title)
	return err
}

func (*Video) GetVideoDaoListById(userId uint64) (PublishList []dao.Video, err error) {
	sqlStr := "select * from video where uploader_id = ? order by id desc"
	err = global.GVAR_SQLX_DB.Select(&PublishList, sqlStr, userId)
	return
}

func (*Video) GetVideoDaoList() (videoDaoList []dao.Video, err error) {
	sqlStr := "select * from video order by id desc limit 0, ?"
	err = global.GVAR_SQLX_DB.Select(&videoDaoList, sqlStr, global.GVAR_FEED_NUM)
	return
}
