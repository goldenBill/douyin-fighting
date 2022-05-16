package dao

import (
	"github.com/jmoiron/sqlx"
	"time"
)

var DataBase *sqlx.DB

type VideoDao struct {
	Id            int64     `db:"id"`
	VideoName     string    `db:"video_name"`
	VideoLocation string    `db:"video_location"`
	CoverLocation string    `db:"cover_location"`
	UploaderId    int64     `db:"uploader_id"`
	CreateTime    time.Time `db:"create_time"`
	Description   *string   `db:"description"`
}

type UserDao struct {
	Id         int64     `db:"id"`
	Name       string    `db:"name"`
	Password   string    `db:"password"`
	Token      string    `db:"token"`
	CreateTime time.Time `db:"create_time"`
}
