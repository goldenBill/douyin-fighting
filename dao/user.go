package dao

import (
	"time"
)

type User struct {
	ID             uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserID         uint64    `gorm:"column:user_id;NOT NULL"`
	Name           string    `gorm:"column:name;NOT NULL"`
	Password       string    `gorm:"column:password;NOT NULL"`
	FollowCount    int64     `gorm:"column:follow_count;NOT NULL"`
	FollowerCount  int64     `gorm:"column:follower_count;NOT NULL"`
	TotalFavorited int64     `gorm:"column:total_favorited;NOT NULL"`
	FavoriteCount  int64     `gorm:"column:favorite_count;NOT NULL"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ExtInfo        *string   `gorm:"column:ext_info"`
}

func (User) TableName() string {
	return "users"
}
