package service

type void struct{}

var member void

// Redis 中 key 的模板
var (
	Header               = ""
	UserPattern          = "user:%d"
	UserFavoritePattern  = "favorite:%d"
	CelebrityPattern     = "celebrity:%d"
	FollowerPattern      = "follower:%d"
	VideoPattern         = "Video:%d"
	CommentPattern       = "Comment:%d"
	VideoCommentsPattern = "CommentsOfVideo:%d"
	PublishPattern       = "Publish:%d"
	EmptyPattern         = "Empty:%d"
)

// VideoFavoriteCountAPI 接收视频喜欢数目的 api 结构体
type VideoFavoriteCountAPI struct {
	VideoID       uint64
	FavoriteCount int64
}
