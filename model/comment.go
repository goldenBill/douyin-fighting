package model

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	CommentID uint64         `gorm:"column:comment_id;primary_key;NOT NULL" redis:"-"`
	VideoID   uint64         `gorm:"column:video_id;index:video_user,priority:1;NOT NULL" redis:"video_id"`
	UserID    uint64         `gorm:"column:user_id;index:video_user,priority:2;NOT NULL" redis:"user_id"`
	Content   string         `gorm:"content:content;NOT NULL" redis:"content"`
	CreatedAt time.Time      `gorm:"column:created_at" redis:"-"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" redis:"-"`
}
