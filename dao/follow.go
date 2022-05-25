package dao

import "time"

type Follow struct {
	ID        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserID    uint64    `gorm:"column:user_id;NOT NULL"`
	FollowId  uint64    `gorm:"column:follow_id;NOT NULL"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:true;NOT NULL"`
}

func (Follow) TableName() string {
	return "follows"
}
