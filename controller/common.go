package controller

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

type User struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	FollowCount   uint64 `json:"follow_count,omitempty"`
	FollowerCount uint64 `json:"follower_count,omitempty"`
	IsFollow      bool   `json:"is_follow"`
}

type Video struct {
	ID            uint64 `json:"id"`
	Author        User   `json:"author"`
	PlayUrl       string `json:"play_url"`
	CoverUrl      string `json:"cover_url"`
	FavoriteCount uint64 `json:"favorite_count"`
	CommentCount  uint64 `json:"comment_count"`
	IsFavorite    bool   `json:"is_favorite"`
	Title         string `json:"title"`
}

type Comment struct {
	ID         uint64 `json:"id,omitempty"`
	User       User   `json:"user"`
	Content    string `json:"content,omitempty"`
	CreateDate string `json:"create_date,omitempty"`
}
