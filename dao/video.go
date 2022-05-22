package dao

import "time"

type Video struct {
	ID            uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	VideoID       uint64    `gorm:"column:video_id;NOT NULL"`
	Title         string    `gorm:"column:title;NOT NULL"`
	PlayUrl       string    `gorm:"column:play_url;NOT NULL"`
	CoverUrl      string    `gorm:"column:cover_url;NOT NULL"`
	FavoriteCount uint64    `gorm:"column:favorite_count;NOT NULL"`
	CommentCount  uint64    `gorm:"column:comment_count;NOT NULL"`
	UserID        uint64    `gorm:"column:user_id;NOT NULL"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime:true;NOT NULL"`
	Active        bool      `gorm:"column:active;NOT NULL"`
	ExtInfo       *string   `gorm:"column:ext_info"`
}

func (Video) TableName() string {
	return "videos"
}
