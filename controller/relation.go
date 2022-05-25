package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/service"
	"log"
	"net/http"
	"strconv"
)

type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

//// FollowActionRequest 点赞操作请求结构体
//type FollowActionRequest struct {
//	UserID     uint64 `form:"user_id" json:"user_id"`
//	Token      string `form:"token" json:"token"`
//	ToUserId   uint64 `form:"to_user_id" json:"to_user_id"`
//	ActionType uint   `form:"action_type" json:"action_type"`
//}

// RelationAction no practical effect, just check if token is valid
func RelationAction(c *gin.Context) {
	userId := c.GetUint64("UserID")
	toUserId, _ := strconv.ParseUint(c.Query("to_user_id"), 10, 64)

	log.Println("userid: ", userId)
	log.Println("touserid: ", toUserId)

	isFollow := service.IsFollow(userId, toUserId)

	if isFollow {
		//已关注，取消关注
		log.Println("action:取消关注")
		service.RemoveFollow(userId, toUserId)
		c.JSON(http.StatusOK, Response{StatusCode: 0})
	} else {
		//未关注，关注
		log.Println("action:关注")
		service.AddFollow(userId, toUserId)
		c.JSON(http.StatusOK, Response{StatusCode: 0})
	}
}

// FollowList all users have same follow list
func FollowList(c *gin.Context) {
	userId := c.GetUint64("UserID")
	log.Println("userId: ", userId)
	follows := service.GetFollow(userId)
	var backUser []User
	for _, follow := range follows {
		tempUser, _ := service.UserInfoByUserID(follow.FollowId)
		backUser = append(backUser, User{
			ID:            tempUser.UserID,
			Name:          tempUser.Name,
			FollowCount:   tempUser.FollowCount,
			FollowerCount: tempUser.FollowerCount,
			IsFollow:      true,
		})
	}
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: backUser,
	})

}

// FollowerList all users have same follower list
func FollowerList(c *gin.Context) {
	userId := c.GetUint64("UserID")
	fans := service.GetFans(userId)
	var backUser []User
	for _, fan := range fans {
		tempUser, _ := service.UserInfoByUserID(fan.UserID)
		backUser = append(backUser, User{
			ID:            tempUser.UserID,
			Name:          tempUser.Name,
			FollowCount:   tempUser.FollowCount,
			FollowerCount: tempUser.FollowerCount,
			IsFollow:      service.IsFollow(userId, tempUser.UserID),
		})
	}
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: backUser,
	})
}
