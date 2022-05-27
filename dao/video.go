package dao

import "time"

type Video struct {
	VideoID       uint64    `gorm:"column:video_id;primary_key;NOT NULL"`
	Title         string    `gorm:"column:title;NOT NULL"`
	PlayName      string    `gorm:"column:play_name;NOT NULL"`
	CoverName     string    `gorm:"column:cover_name;NOT NULL"`
	FavoriteCount int64     `gorm:"column:favorite_count;NOT NULL"`
	CommentCount  int64     `gorm:"column:comment_count;NOT NULL"`
	AuthorID      uint64    `gorm:"column:author_id;NOT NULL"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	ExtInfo       *string   `gorm:"column:ext_info"`
}

func (Video) TableName() string {
	return "videos"
}
