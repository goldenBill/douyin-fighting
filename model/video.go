package model

import (
	initialize "simple-demo/initialize"
	"time"
)

type Video struct {
	Id            int64     `json:"id,omitempty"`
	AuthId        int64     `json:"auth_id"`
	PlayUrl       string    `json:"play_url" json:"play_url,omitempty"`
	CoverUrl      string    `json:"cover_url,omitempty"`
	Description   string    `json:"description"`
	FavoriteCount int64     `json:"favorite_count,omitempty"`
	CommentCount  int64     `json:"comment_count,omitempty"`
	LatestTime    time.Time `json:"latest_time"`
}

func GetAllVideo() []Video {
	var videoList []Video
	initialize.GLOBAL_DB.Find(&videoList)
	return videoList
}

func GetVideoByUser(user User) []Video {
	var videoList []Video
	initialize.GLOBAL_DB.Where(&Video{
		AuthId: user.Id,
	}).Find(&videoList)
	return videoList
}

func InsertVideo(video Video) {
	initialize.GLOBAL_DB.Create(&video)
}
