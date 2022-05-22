package controller

import (
	"github.com/gin-gonic/gin"
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

// Feed : same demo video list for every request
func Feed(c *gin.Context) {
	//// token 验证
	//var tokenString = c.DefaultQuery("token", "")
	//if tokenString != "" {
	//	_, err := service.ParseToken(tokenString)
	//	if err != nil {
	//		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
	//		return
	//	}
	//}
	defaultTime := strconv.FormatInt(time.Now().UnixMilli(), 10)
	latestTimeUnix, _ := strconv.ParseInt(c.DefaultQuery("latest_time", defaultTime), 10, 64)
	latestTime := time.UnixMilli(latestTimeUnix)
	videoDaoList, rows := service.GetVideos(latestTime)
	if rows == 0 {
		c.JSON(http.StatusOK, FeedResponse{
			Response:  Response{StatusCode: 1, StatusMsg: "no latest video"},
			VideoList: nil,
			NextTime:  latestTime.UnixMilli(),
		})
	}

	var videoList []Video
	for _, videoDao := range videoDaoList {
		VideoLocation := "./public/" + videoDao.PlayUrl
		if _, err := os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := "./public/" + videoDao.CoverUrl
		if _, err := os.Stat(CoverLocation); err != nil {
			continue
		}
		userDao, _ := service.UserInfoByUserID(videoDao.UserID)
		var author = User{
			ID:            userDao.UserID,
			Name:          userDao.Name,
			FollowCount:   userDao.FollowCount,
			FollowerCount: userDao.FollowerCount,
			IsFollow:      false,
		}
		video := Video{
			ID:            videoDao.VideoID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static" + videoDao.PlayUrl,
			CoverUrl:      "http://" + c.Request.Host + "/static" + videoDao.CoverUrl,
			FavoriteCount: videoDao.FavoriteCount,
			CommentCount:  videoDao.CommentCount,
			Title:         videoDao.Title,
			IsFavorite:    false,
		}
		videoList = append(videoList, video)
	}

	nextTime := latestTime
	if len(videoDaoList) > 0 {
		nextTime = videoDaoList[len(videoDaoList)-1].CreatedAt
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "OK"},
		VideoList: videoList,
		NextTime:  nextTime.UnixMilli(),
	})
}
