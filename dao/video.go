package dao

import "time"

type Video struct {
	VideoID       uint64    `gorm:"column:video_id;primary_key;NOT NULL" redis:"-"`
	Title         string    `gorm:"column:title;NOT NULL" redis:"title"`
	PlayName      string    `gorm:"column:play_name;NOT NULL" redis:"play_name"`
	CoverName     string    `gorm:"column:cover_name;NOT NULL" redis:"cover_name"`
	FavoriteCount int64     `gorm:"column:favorite_count;NOT NULL" redis:"favorite_count"`
	CommentCount  int64     `gorm:"column:comment_count;NOT NULL" redis:"comment_count"`
	AuthorID      uint64    `gorm:"column:author_id;NOT NULL" redis:"author_id"`
	CreatedAt     time.Time `gorm:"column:created_at" redis:"-"`
	ExtInfo       *string   `gorm:"column:ext_info" redis:"-"`
}

func (Video) TableName() string {
	return "videos"
}
