package dao

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID        uint64         `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	CommentID uint64         `gorm:"column:comment_id;NOT NULL"`
	VideoID   uint64         `gorm:"column:video_id;NOT NULL;index"`
	UserID    uint64         `gorm:"column:user_id;NOT NULL"`
	Content   string         `gorm:"content:user_id;NOT NULL"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Comment) TableName() string {
	return "comments"
}
