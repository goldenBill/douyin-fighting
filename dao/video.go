package dao

import "time"

type Video struct {
	Id            uint64    `db:"id"`
	VideoName     string    `db:"video_name"`
	VideoLocation string    `db:"video_location"`
	CoverLocation string    `db:"cover_location"`
	UploaderId    uint64    `db:"uploader_id"`
	CreateTime    time.Time `db:"create_time"`
	Title         *string   `db:"title"`
}
