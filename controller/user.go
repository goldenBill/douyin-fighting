package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"net/http"
	"strconv"
)

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin
var usersLoginInfo = map[string]User{
	"zhangleidouyin": {
		Id:            1,
		Name:          "zhanglei",
		FollowCount:   10,
		FollowerCount: 5,
		IsFollow:      true,
	},
}

var userIdSequence = int64(1)

type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Response
	User User `json:"user"`
}

func GetUserInfoByName(username string) *dao.UserDao {
	var userDao dao.UserDao
	sqlStr := "select * from user where name = ?"
	if err := dao.DataBase.Get(&userDao, sqlStr, username); err != nil {
		fmt.Println("exec failed, ", err)
		return nil
	}
	return &userDao
}

func GetUserInfoById(userId int64) *dao.UserDao {
	var userDao dao.UserDao
	sqlStr := "select * from user where id = ?"
	if err := dao.DataBase.Get(&userDao, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return nil
	}
	return &userDao
}

func AddNewUser(username string, password string, token string) int64 {
	sqlStr := "insert into user (name, password, token) values (?, ?, ?)"
	r, err := dao.DataBase.Exec(sqlStr, username, password, token)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return -1
	}
	id, err := r.LastInsertId()
	if err != nil {
		fmt.Println("exec failed, ", err)
		return -1
	}
	return id
}

func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	userDao := GetUserInfoByName(username)
	if password == "" {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "Invalid password"},
		})
	} else if userDao != nil {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User already exist"},
		})
	} else {
		token := username + password
		id := AddNewUser(username, password, token)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0, StatusMsg: "OK"},
			UserId:   id,
			Token:    token,
		})
	}
}

func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	userDao := GetUserInfoByName(username)

	if userDao == nil {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User doesn't exist"},
		})
	} else if userDao.Password != password {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "Wrong password"},
		})
	} else {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0, StatusMsg: "OK"},
			UserId:   userDao.Id,
			Token:    userDao.Token,
		})
	}
}

func UserInfo(c *gin.Context) {
	userId, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	//token := c.Query("token")

	if userDao := GetUserInfoById(userId); userDao == nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User doesn't exist"},
		})
	} else {
		isFollow := false
		user := User{
			Id:            userDao.Id,
			Name:          userDao.Name,
			FollowCount:   GetFollowCount(userDao.Id),
			FollowerCount: GetFollowerCount(userDao.Id),
			IsFollow:      isFollow,
		}
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "OK"},
			User:     user,
		})
	}
}
