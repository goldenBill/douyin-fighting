package dao

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	CommentID uint64         `gorm:"column:comment_id;primary_key;NOT NULL"`
	VideoID   uint64         `gorm:"column:video_id;NOT NULL;index"`
	UserID    uint64         `gorm:"column:user_id;NOT NULL"`
	Content   string         `gorm:"content:content;NOT NULL"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Comment) TableName() string {
	return "comments"
}
