package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/service"
	"net/http"
	"strconv"
)

// FavoriteActionRequest 点赞操作请求结构体
type FavoriteActionRequest struct {
	UserID     uint64 `form:"user_id" json:"user_id"`
	Token      string `form:"token" json:"token"`
	VideoID    uint64 `form:"video_id" json:"video_id"`
	ActionType uint   `form:"action_type" json:"action_type"`
}

// FavoriteListRequest 请求点赞列表结构体
type FavoriteListRequest struct {
	UserID uint64 `form:"user_id"`
	Token  string `form:"token"`
}

// FavoriteAction no practical effect, just check if token is valid
func FavoriteAction(c *gin.Context) {
	// 参数绑定
	var r FavoriteActionRequest
	err := c.ShouldBind(&r)

	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}

	// 判断 action_type 是否正确
	if r.ActionType != 1 && r.ActionType != 2 {
		// action_type 不合法
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "action type error"})
		return
	}

	// 获取 userID
	r.UserID = c.GetUint64("UserID")

	// 判断 video_id 是否正确

	// 点赞操作
	if r.ActionType == 1 {
		err = service.FavoriteAction(r.UserID, r.VideoID)
	} else {
		err = service.CancelFavorite(r.UserID, r.VideoID)
	}
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "server error"})
		return
	}

	c.JSON(http.StatusOK, Response{StatusCode: 0})
}

// FavoriteList all users have same favorite video list
func FavoriteList(c *gin.Context) {
	// 参数绑定
	var r FavoriteListRequest
	err := c.ShouldBind(&r)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "bind error"})
	}

	// 获取 userID
	r.UserID, _ = strconv.ParseUint(c.Query("user_id"), 10, 64)

	videoDaoList, err := service.GetFavoriteListByUserID(r.UserID)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "get favorite list failed"})
		return
	}

	var videoList []Video
	for _, videoDao := range videoDaoList {
		var followCount = service.GetFollowCount(videoDao.AuthorID)
		var followerCount = service.GetFollowerCount(videoDao.AuthorID)
		userDao, err := service.UserInfoByUserID(videoDao.AuthorID)
		if err != nil {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
			return
		}
		var name = userDao.Name
		var isFollow = false
		var author = User{
			ID:            videoDao.AuthorID,
			Name:          name,
			FollowCount:   followCount,
			FollowerCount: followerCount,
			IsFollow:      isFollow,
		}
		var favoriteCount = service.GetFavoriteCount(videoDao.ID)
		var isFavorite = service.GetFavoriteStatus(r.UserID, videoDao.ID)
		video := Video{
			ID:            videoDao.ID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static/video/" + videoDao.PlayName,
			CoverUrl:      "http://" + c.Request.Host + "/static/cover/" + videoDao.CoverName,
			FavoriteCount: favoriteCount,
			CommentCount:  0,
			IsFavorite:    isFavorite,
		}
		videoList = append(videoList, video)
	}

	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "OK",
		},
		VideoList: videoList,
	})
}
