package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

// RelationAction no practical effect, just check if token is valid
func RelationAction(c *gin.Context) {
	c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "Unrealized"})
}

// FollowList all users have same follow list
func FollowList(c *gin.Context) {
	c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "Unrealized"})
}

// FollowerList all users have same follower list
func FollowerList(c *gin.Context) {
	c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "Unrealized"})
}
