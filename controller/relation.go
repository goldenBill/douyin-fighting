package controller

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"simple-demo/model"
	"strconv"
)

type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

// RelationAction no practical effect, just check if token is valid
func RelationAction(c *gin.Context) {
	token := c.Query("token")
	user := model.FindUserByToken(token)
	userId := user.Id
	toUserId, _ := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	log.Println("userid: ", userId)
	log.Println("touserid: ", toUserId)
	isfollow := model.IsFollow(userId, toUserId)
	if isfollow {
		log.Println("action:取消关注")
		model.RemoveFollow(userId, toUserId)
		c.JSON(http.StatusOK, Response{StatusCode: 0})
	} else {
		//未关注，action:关注
		log.Println("action:关注")
		model.AddFollow(userId, toUserId)
		c.JSON(http.StatusOK, Response{StatusCode: 0})
	}

	//token := c.Query("token")
	//
	//if _, exist := usersLoginInfo[token]; exist {
	//	c.JSON(http.StatusOK, Response{StatusCode: 0})
	//} else {
	//	c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
	//}
}

// FollowList all users have same follow list
func FollowList(c *gin.Context) {
	token := c.Query("token")
	user := model.FindUserByToken(token)
	follows := model.GetFollow(user)
	log.Println("follows: ", follows)
	var backUser []User
	for _, follow := range follows {
		tempUser := model.FindUserById(follow.FollowId)
		backUser = append(backUser, User{
			Id:            tempUser.Id,
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

	//c.JSON(http.StatusOK, UserListResponse{
	//	Response: Response{
	//		StatusCode: 0,
	//	},
	//	UserList: []User{DemoUser},
	//})
}

// FollowerList all users have same follower list
func FollowerList(c *gin.Context) {
	token := c.Query("token")
	user := model.FindUserByToken(token)
	fans := model.GetFans(user)
	log.Println("fans: ", fans)
	var backUser []User
	for _, fan := range fans {
		tempUser := model.FindUserById(fan.UserId)

		backUser = append(backUser, User{
			Id:            tempUser.Id,
			Name:          tempUser.Name,
			FollowCount:   tempUser.FollowCount,
			FollowerCount: tempUser.FollowerCount,
			IsFollow:      model.IsFollow(user.Id, tempUser.Id),
		})
	}

	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: backUser,
	})

	//c.JSON(http.StatusOK, UserListResponse{
	//	Response: Response{
	//		StatusCode: 0,
	//	},
	//	UserList: []User{DemoUser},
	//})
}
