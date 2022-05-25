package dao

import (
	"time"
)

type Follow struct {
	ID          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	FollowID    uint64    `gorm:"column:follow_id;NOT NULL"`
	CelebrityID uint64    `gorm:"column:celebrity_id;NOT NULL"`
	FollowerID  uint64    `gorm:"column:follower_id;NOT NULL"`
	IsFollow    bool      `gorm:"column:is_follow;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (Follow) TableName() string {
	return "follows"
}
