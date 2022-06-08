package model

import (
	"time"
)

type User struct {
	UserID         uint64    `gorm:"column:user_id;primary_key;NOT NULL" redis:"user_id"`
	Name           string    `gorm:"column:name;NOT NULL" redis:"name"`
	Password       string    `gorm:"column:password;NOT NULL" redis:"password"`
	FollowCount    int64     `gorm:"-" redis:"follow_count"`
	FollowerCount  int64     `gorm:"-" redis:"follower_count"`
	TotalFavorited int64     `gorm:"-" redis:"total_favorited"`
	FavoriteCount  int64     `gorm:"-" redis:"favorite_count"`
	CreatedAt      time.Time `gorm:"column:created_at" redis:"-"`
	ExtInfo        *string   `gorm:"column:ext_info" redis:"-"`
}
