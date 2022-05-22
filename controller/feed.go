package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	latestTimeUnix, _ := strconv.ParseInt(c.DefaultQuery("latest_time", "1652480260 * 1000"), 10, 64)
	//var token = c.DefaultQuery("token", "")

	videoDaoList, _ := videoService.GetVideoDaoList()

	var videoList []Video
	for _, videoDao := range videoDaoList {
		//if videoDao.CreateTime.UnixNano()/1e6 < latestTimeUnix {
		//	fmt.Println(videoDao.CreateTime.Unix(), latestTimeUnix)
		//	continue
		//}
		userDao, _ := userService.UserInfoById(videoDao.UploaderId)
		var author = User{
			Id:            userDao.UserId,
			Name:          userDao.Name,
			FollowCount:   0,
			FollowerCount: 0,
			IsFollow:      false,
		}
		video := Video{
			Id:            videoDao.Id,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static" + videoDao.VideoLocation + videoDao.VideoName,
			CoverUrl:      "http://" + c.Request.Host + "/static" + videoDao.CoverLocation,
			FavoriteCount: 0,
			CommentCount:  0,
			IsFavorite:    false,
			Title:         videoDao.Title,
		}
		videoList = append(videoList, video)
	}

	nextTime := latestTimeUnix
	if len(videoDaoList) > 0 {
		nextTime = videoDaoList[len(videoDaoList)-1].CreateTime.UnixNano() / 1e6
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "OK"},
		VideoList: videoList,
		NextTime:  nextTime,
	})
}
