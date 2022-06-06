package service

type void struct{}

var member void

var (
	Header               string = ""
	UserPattern          string = "user:%d"
	UserFavoritePattern  string = "favorite:%d"
	CelebrityPattern     string = "celebrity:%d"
	FollowerPattern      string = "follower:%d"
	VideoPattern         string = "Video:%d"
	CommentPattern       string = "Comment:%d"
	VideoCommentsPattern string = "CommentsOfVideo:%d"
	PublishPattern              = "Publish:%d"
)

type VideoFavoriteCountAPI struct {
	VideoID       uint64
	FavoriteCount int64
}
