package controller

// Response 基础响应结构体
type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

// User 用户信息响应结构体
type User struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	FollowCount    int64  `json:"follow_count,omitempty"`
	FollowerCount  int64  `json:"follower_count,omitempty"`
	TotalFavorited int64  `json:"total_favorited,omitempty"`
	FavoriteCount  int64  `json:"favorite_count,omitempty"`
	IsFollow       bool   `json:"is_follow"`
}

// Video 视频信息响应结构体
type Video struct {
	ID            uint64 `json:"id"`
	Author        User   `json:"author"`
	PlayUrl       string `json:"play_url"`
	CoverUrl      string `json:"cover_url"`
	FavoriteCount int64  `json:"favorite_count"`
	CommentCount  int64  `json:"comment_count"`
	IsFavorite    bool   `json:"is_favorite"`
	Title         string `json:"title"`
}

// Comment 评论信息响应结构体
type Comment struct {
	ID         uint64 `json:"id,omitempty"`
	User       User   `json:"user"`
	Content    string `json:"content,omitempty"`
	CreateDate string `json:"create_date,omitempty"`
}
