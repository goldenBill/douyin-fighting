package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/service"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"strconv"
)

// UserListResponse 关注列表或粉丝列表请求结构体
type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

// RelationAction 评论操作
func RelationAction(c *gin.Context) {
	// 参数绑定
	toUserID, _ := strconv.ParseUint(c.Query("to_user_id"), 10, 64)
	actionType, _ := strconv.ParseInt(c.Query("action_type"), 10, 64)
	// 判断 action_type 是否正确
	if actionType != 1 && actionType != 2 {
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "action type error"})
		return
	}
	// 获取当前用户的 ID
	viewID := c.GetUint64("UserID")
	// 关注操作
	if actionType == 1 {
		if err := service.AddFollow(viewID, toUserID); err != nil {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "server error"})
			return
		}
	} else {
		if err := service.CancelFollow(viewID, toUserID); err != nil {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "server error"})
			return
		}
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "OK"})
}

// FollowList 获取关注列表
func FollowList(c *gin.Context) {
	// 参数绑定
	followerID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "request is invalid"})
		return
	}
	// 判断是否登录
	var (
		isLogin  bool
		viewerID uint64
	)
	// 判断传入的token是否合法，用户是否存在
	if token := c.Query("token"); token != "" {
		claims, err := util.ParseToken(token)
		if err == nil {
			viewerID = claims.UserID
			isLogin = true
		}
	}
	// 获取用户的关注列表
	celebrityList, err := service.GetFollowListByUserID(followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "get Follower list failed"})
		return
	}
	// 生成 response 数据
	celebrityIDList := make([]uint64, len(celebrityList))
	var userList []User
	for idx, celebrity := range celebrityList {
		var user = User{
			ID:            celebrity.UserID,
			Name:          celebrity.Name,
			FollowCount:   celebrity.FollowCount,
			FollowerCount: celebrity.FollowerCount,
		}
		userList = append(userList, user)
		celebrityIDList[idx] = celebrity.UserID
	}
	// 批量处理
	if isLogin {
		// 登录时，获取是否关注，否则总是为false
		isFollowList, _ := service.GetFollowStatusList(viewerID, celebrityIDList)
		for idx, isFollow := range isFollowList {
			userList[idx].IsFollow = isFollow
		}
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{StatusCode: 0, StatusMsg: "OK"},
		UserList: userList,
	})
}

// FollowerList 获取粉丝列表
func FollowerList(c *gin.Context) {
	// 参数绑定
	celebrityID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	// 判断是否登录
	var (
		isLogin  bool
		viewerID uint64
	)
	// 判断传入的token是否合法，用户是否存在
	if token := c.Query("token"); token != "" {
		claims, err := util.ParseToken(token)
		if err == nil {
			viewerID = claims.UserID
			isLogin = true
		}
	}
	// 获取用户的粉丝列表
	followerList, err := service.GetFollowerListByUserID(celebrityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "get Follower list failed"})
		return
	}
	// 生成 response 数据
	var userList []User
	for _, follower := range followerList {
		isFollow := false
		if isLogin {
			// 登录时，获取是否关注，否则总是为false
			isFollow, _ = service.GetFollowStatus(viewerID, follower.UserID)
		}
		var user = User{
			ID:            follower.UserID,
			Name:          follower.Name,
			FollowCount:   follower.FollowCount,
			FollowerCount: follower.FollowerCount,
			IsFollow:      isFollow,
		}
		userList = append(userList, user)
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{StatusCode: 0, StatusMsg: "OK"},
		UserList: userList,
	})
}
