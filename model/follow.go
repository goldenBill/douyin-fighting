package model

import (
	"time"
)

type Follow struct {
	FollowID    uint64    `gorm:"column:follow_id;primary_key;NOT NULL"`
	CelebrityID uint64    `gorm:"column:celebrity_id;NOT NULL;index:idx_01,priority:2;index:idx_02"`
	FollowerID  uint64    `gorm:"column:follower_id;NOT NULL;index:idx_01,priority:1"`
	IsFollow    bool      `gorm:"column:is_follow;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}
