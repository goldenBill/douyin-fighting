package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/service"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"regexp"
	"strconv"
	"unicode/utf8"
)

// UserLoginResponse 用户登录响应结构体
type UserLoginResponse struct {
	Response
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
}

// UserResponse 用户信息响应结构体
type UserResponse struct {
	Response
	User User `json:"user"`
}

// Register 用户注册账号
func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	// 验证用户名合法性
	if utf8.RuneCountInString(username) > global.MAX_USERNAME_LENGTH ||
		utf8.RuneCountInString(username) <= 0 {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "非法用户名"})
		return
	}
	// 验证密码合法性
	if ok, _ := regexp.MatchString(global.MIN_PASSWORD_PATTERN, password); !ok {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "密码长度6-32，由字母大小写下划线组成"})
		return
	}
	// 注册用户到数据库
	userModel, err := service.Register(username, password)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	// 生成对应 token
	tokenString, err := util.GenerateToken(userModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, UserLoginResponse{
		Response: Response{StatusCode: 0, StatusMsg: "OK"},
		UserID:   userModel.UserID,
		Token:    tokenString,
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	// 从数据库查询用户信息
	userModel, err := service.Login(username, password)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "用户名或密码错误"})
		return
	}
	// 生成对应 token
	tokenString, err := util.GenerateToken(userModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, UserLoginResponse{
		Response: Response{StatusCode: 0, StatusMsg: "OK"},
		UserID:   userModel.UserID,
		Token:    tokenString,
	})
}

// UserInfo 获取用户信息
func UserInfo(c *gin.Context) {
	// 获取指定用户的 ID
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "request is invalid"})
		return
	}
	userModel, err := service.UserInfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	// 获取当前用户的 ID
	viewerID := c.GetUint64("UserID")
	// 查询当前用户是否关注指定用户
	isFollow, err := service.GetFollowStatus(viewerID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	// 返回成功并生成响应 json
	c.JSON(http.StatusOK, UserResponse{
		Response: Response{StatusCode: 0, StatusMsg: "OK"},
		User: User{
			ID:             userID,
			Name:           userModel.Name,
			FollowCount:    userModel.FollowCount,
			FollowerCount:  userModel.FollowerCount,
			TotalFavorited: userModel.TotalFavorited,
			FavoriteCount:  userModel.FavoriteCount,
			IsFollow:       isFollow,
		},
	})
}
