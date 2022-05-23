package dao

type Video struct {
	ID            uint64 `db:"id" gorm:"PrimaryKey"`
	VideoID       uint64 `db:"video_id"`
	Title         string `db:"title"`
	PlayUrl       string `db:"play_url"`
	CoverUrl      string `db:"cover_url"`
	FavoriteCount int64  `db:"favorite_count"`
	CommentCount  int64  `db:"comment_count"`
	UserID        uint64 `db:"user_id"`
	CreatedAt     int64  `db:"created_at"`
	Active        bool   `db:"active"`
}

func (Video) TableName() string {
	return "videos"
}
