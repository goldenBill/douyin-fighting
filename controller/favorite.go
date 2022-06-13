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

// FavoriteListRequest 点赞列表请求结构体
type FavoriteListRequest struct {
	UserID uint64 `form:"user_id"`
	Token  string `form:"token"`
}

// FavoriteAction 点赞操作
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
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "action type error"})
		return
	}
	// 获取当前用户的 ID
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
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, Response{StatusCode: 0})
}

// FavoriteList 返回用户喜欢的视频列表
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
		}
	}
	// 获取用户的点赞列表
	videoModelList, err := service.GetFavoriteListByUserID(r.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "get favorite list failed"})
		return
	}
	// 产生相应结构体
	celebrityIDList := make([]uint64, len(videoModelList))
	videoIDList := make([]uint64, len(videoModelList))
	var videoList []Video
	for idx, each := range videoModelList {
		userModel, err := service.UserInfoByUserID(each.AuthorID)
		if err != nil {
			continue
		}
		var author = User{
			ID:             each.AuthorID,
			Name:           userModel.Name,
			FollowCount:    userModel.FollowCount,
			FollowerCount:  userModel.FollowerCount,
			TotalFavorited: userModel.TotalFavorited,
			FavoriteCount:  userModel.FavoriteCount,
		}
		var isFavorite bool // 是否对视频点赞
		video := Video{
			ID:            each.VideoID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static/video/" + each.PlayName,
			CoverUrl:      "http://" + c.Request.Host + "/static/cover/" + each.CoverName,
			FavoriteCount: each.FavoriteCount,
			CommentCount:  each.CommentCount,
			IsFavorite:    isFavorite,
		}
		videoList = append(videoList, video)
		celebrityIDList[idx] = each.AuthorID
		videoIDList[idx] = each.VideoID
	}
	// 批量处理
	if isLogin {
		// 登录时，获取是否关注以及是否点赞，否则总是为false
		isFollowList, _ := service.GetFollowStatusList(userID, celebrityIDList)
		isFavoriteList, _ := service.GetFavoriteStatusList(userID, videoIDList)
		for i := 0; i < len(videoModelList); i++ {
			videoList[i].Author.IsFollow = isFollowList[i]
			videoList[i].IsFavorite = isFavoriteList[i]
		}
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "OK",
		},
		VideoList: videoList,
	})
}
