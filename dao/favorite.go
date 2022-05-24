package dao

import (
	"time"
)

type Favorite struct {
	ID         uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	FavoriteID uint64    `gorm:"column:favorite_id;NOT NULL"`
	VideoID    uint64    `gorm:"column:video_id;NOT NULL"`
	UserID     uint64    `gorm:"column:user_id;NOT NULL"`
	IsFavorite bool      `gorm:"column:is_favorite;NOT NULL"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}
