package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"simple-demo/model"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	token := c.Query("token")
	user := model.FindUserByToken(token)
	//数据库video
	videoList := model.GetAllVideo()
	//url video
	var backVideo []Video
	for _, video := range videoList {
		tempUser := model.FindUserById(video.AuthId)
		tempVideo := Video{
			Id: video.Id,
			Author: User{
				Id:            tempUser.Id,
				Name:          tempUser.Name,
				FollowCount:   int64(len(model.GetFollow(tempUser))),
				FollowerCount: int64(len(model.GetFans(tempUser))),
				IsFollow:      model.IsFollow(user.Id, tempUser.Id),
			},
			PlayUrl:       video.PlayUrl,
			CoverUrl:      video.CoverUrl,
			FavoriteCount: video.FavoriteCount,
			CommentCount:  video.CommentCount,
			IsFavorite:    false,
		}
		backVideo = append(backVideo, tempVideo)
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: backVideo,
		NextTime:  time.Now().Unix(),
	})

	//c.JSON(http.StatusOK, FeedResponse{
	//	Response:  Response{StatusCode: 0},
	//	VideoList: DemoVideos,
	//	NextTime:  time.Now().Unix(),
	//})
}
