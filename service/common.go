package service

type void struct{}

var member void

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

type VideoFavoriteCountAPI struct {
	VideoID       uint64
	FavoriteCount int64
}
