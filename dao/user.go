package dao

import (
	"time"
)

type User struct {
	ID            uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserID        uint64    `gorm:"column:user_id;NOT NULL"`
	Name          string    `gorm:"column:name;NOT NULL"`
	Password      string    `gorm:"column:password;NOT NULL"`
	FollowCount   uint64    `gorm:"column:follow_count;default:0;NOT NULL"`
	FollowerCount uint64    `gorm:"column:follower_count;default:0;NOT NULL"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime:true;NOT NULL"`
	ExtInfo       *string   `gorm:"column:ext_info"`
}

func (User) TableName() string {
	return "users"
}
