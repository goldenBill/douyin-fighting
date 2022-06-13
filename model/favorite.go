package model

import (
	"time"
)

type Favorite struct {
	FavoriteID uint64    `gorm:"column:favorite_id;primary_key;NOT NULL"`
	VideoID    uint64    `gorm:"column:video_id;NOT NULL;index:idx_01,priority:2;index:idx_02"`
	UserID     uint64    `gorm:"column:user_id;NOT NULL;index:idx_01,priority:1"`
	IsFavorite bool      `gorm:"column:is_favorite;NOT NULL"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}
