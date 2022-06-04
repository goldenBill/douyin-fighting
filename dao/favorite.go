package dao

import (
	"time"
)

type Favorite struct {
	FavoriteID uint64    `gorm:"column:favorite_id;primary_key;NOT NULL"`
	VideoID    uint64    `gorm:"column:video_id;NOT NULL"`
	UserID     uint64    `gorm:"column:user_id;NOT NULL"`
	IsFavorite bool      `gorm:"column:is_favorite;NOT NULL"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:deleted_at"`
}
