package dao

import (
	"time"
)

type User struct {
	Id       uint64    `db:"id"`
	UserId   uint64    `db:"user_id"`
	Name     string    `db:"name"`
	Password string    `db:"password"`
	CreateAt time.Time `db:"create_at" gorm:"autoCreateTime:true"`
	ExtInfo  *string   `db:"ext_info" gorm:"-"`
}

func (User) TableName() string {
	return "user"
}
