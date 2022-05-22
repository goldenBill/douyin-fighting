package dao

import (
	"time"
)

type User struct {
	Id            uint64
	UserId        uint64
	Name          string
	Password      string
	FollowCount   uint64
	FollowerCount uint64
	CreateAt      time.Time ` gorm:"autoCreateTime:true"`
	ExtInfo       *string   ` gorm:"-"`
}

func (User) TableName() string {
	return "user"
}
