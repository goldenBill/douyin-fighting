package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/service"
	"github.com/goldenBill/douyin-fighting/util"
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
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}
	// 判断 action_type 是否正确
	if r.ActionType != 1 && r.ActionType != 2 {
		// action_type 不合法
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "action type error"})
		return
	}
	// 获取 userID
	r.UserID = c.GetUint64("UserID")
	// 点赞操作
	if r.ActionType == 1 {
		err = service.AddFavorite(r.UserID, r.VideoID)
	} else {
		err = service.CancelFavorite(r.UserID, r.VideoID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "server error"})
		return
	}
	//返回成功并生成响应 json
	c.JSON(http.StatusOK, Response{StatusCode: 0})
}

// FavoriteList all users have same favorite video list
func FavoriteList(c *gin.Context) {
	// 参数绑定
	var r FavoriteListRequest
	err := c.ShouldBind(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "bind error"})
	}
	// 获取 userID
	r.UserID, err = strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "request is invalid"})
		return
	}
	// 判断是否登录
	var isLogin bool
	var userID uint64
	// 判断传入的token是否合法，用户是否存在
	if token := c.Query("token"); token != "" {
		claims, err := util.ParseToken(token)
		if err == nil {
			userID = claims.UserID
			isLogin = true
			//if service.IsUserIDExist(userID) {
			//	isLogin = true
			//}
		}
	}
	//获取用户的点赞列表
	videoDaoList, err := service.GetFavoriteListByUserID(r.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "get favorite list failed"})
		return
	}
	//产生相应结构体
	celebrityIDList := make([]uint64, len(videoDaoList))
	videoIDList := make([]uint64, len(videoDaoList))
	var videoList []Video
	for idx, videoDao := range videoDaoList {
		userDao, err := service.UserInfoByUserID(videoDao.AuthorID)
		if err != nil {
			continue
		}
		var author = User{
			ID:             videoDao.AuthorID,
			Name:           userDao.Name,
			FollowCount:    userDao.FollowCount,
			FollowerCount:  userDao.FollowerCount,
			TotalFavorited: userDao.TotalFavorited,
			FavoriteCount:  userDao.FavoriteCount,
		}
		var isFavorite bool // 是否对视频点赞
		video := Video{
			ID:            videoDao.VideoID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static/video/" + videoDao.PlayName,
			CoverUrl:      "http://" + c.Request.Host + "/static/cover/" + videoDao.CoverName,
			FavoriteCount: videoDao.FavoriteCount,
			CommentCount:  videoDao.CommentCount,
			IsFavorite:    isFavorite,
		}
		videoList = append(videoList, video)
		celebrityIDList[idx] = videoDao.AuthorID
		videoIDList[idx] = videoDao.VideoID
	}
	// 批量处理
	if isLogin {
		// 登录时，获取是否关注，否则总是为false
		isFollowList, _ := service.GetFollowStatusList(userID, celebrityIDList)
		isFavoriteList, _ := service.GetFavoriteStatusList(userID, videoIDList)
		for i := 0; i < len(videoDaoList); i++ {
			videoList[i].Author.IsFollow = isFollowList[i]
			videoList[i].IsFavorite = isFavoriteList[i]
		}
	}
	//返回成功并生成响应 json
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "OK",
		},
		VideoList: videoList,
	})
}
