package dao

type Video struct {
	ID            uint64 `db:"id" gorm:"PrimaryKey"`
	VideoID       uint64 `db:"video_id"`
	Title         string `db:"title"`
	PlayName      string `db:"play_name"`
	CoverName     string `db:"cover_name"`
	FavoriteCount int64  `db:"favorite_count"`
	CommentCount  int64  `db:"comment_count"`
	UserID        uint64 `db:"user_id"`
	CreatedAt     int64  `db:"created_at"`
}

func (Video) TableName() string {
	return "videos"
}
