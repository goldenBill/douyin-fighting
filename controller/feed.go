package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/service"
	"net/http"
	"os"
	"path/filepath"
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
	MaxNumVideo := global.GVAR_FEED_NUM
	var videos []dao.Video
	result := service.GetVideos(&videos, LatestTime, MaxNumVideo)
	if result.Error != nil {
		println("???")
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: result.Error.Error()})
		return
	} else if result.RowsAffected == 0 {
		println("!!!!!1")
		c.JSON(http.StatusOK, FeedResponse{
			Response:  Response{StatusCode: 0},
			VideoList: nil,
			NextTime:  LatestTime,
		})
		return
	}

	var videoList []Video
	for _, video_ := range videos {
		VideoLocation := filepath.Join(global.GVAR_VIDEO_ADDR, video_.PlayName)
		if _, err := os.Stat(VideoLocation); err != nil {
			println("####333", VideoLocation)
			continue
		}
		CoverLocation := filepath.Join(global.GVAR_COVER_ADDR, video_.CoverName)
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
			PlayUrl:       "http://" + c.Request.Host + "/static/video/" + video_.PlayName,
			CoverUrl:      "http://" + c.Request.Host + "/static/public/" + video_.CoverName,
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
