package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/service"
	"net/http"
	"os"
	"strconv"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	var CurrentTimeInt int64 = time.Now().UnixMilli()
	var CurrentTime string = strconv.FormatInt(CurrentTimeInt, 10)
	var LatestTimeStr string = c.DefaultQuery("latest_time", CurrentTime)
	LatestTime, _ := strconv.ParseInt(LatestTimeStr, 10, 64)
	MaxNumVideo := 30
	var videos []dao.Video
	if rows := service.GetVideos(&videos, LatestTime, MaxNumVideo); rows == 0 {
		c.JSON(http.StatusOK, FeedResponse{
			Response:  Response{StatusCode: 0},
			VideoList: nil,
			NextTime:  LatestTime,
		})
		return
	}

	var videoList []Video
	for _, video_ := range videos {
		VideoLocation := "./public/" + video_.PlayUrl
		if _, err := os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := "./public/" + video_.CoverUrl
		if _, err := os.Stat(CoverLocation); err != nil {
			continue
		}
		author_, err := service.UserInfoByUserID(video_.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		isFollow := false
		author := User{
			ID:            author_.UserID,
			Name:          author_.Name,
			FollowCount:   author_.FollowerCount,
			FollowerCount: author_.FollowerCount,
			IsFollow:      isFollow,
		}
		isFavorite := false
		video := Video{
			ID:            video_.VideoID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static/" + video_.PlayUrl,
			CoverUrl:      "http://" + c.Request.Host + "/static/" + video_.CoverUrl,
			FavoriteCount: video_.FavoriteCount,
			CommentCount:  video_.CommentCount,
			Title:         video_.Title,
			IsFavorite:    isFavorite,
		}
		videoList = append(videoList, video)
	}

	nextTime := videos[len(videos)-1].CreatedAt

	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: videoList,
		NextTime:  nextTime,
	})
}
