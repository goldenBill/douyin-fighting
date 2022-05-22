package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"path/filepath"
	"strconv"
)

type VideoListResponse struct {
	Response
	VideoList []Video `json:"video_list"`
}

// Publish : check token then save upload file to public directory
func Publish(c *gin.Context) {
	tokenString := c.PostForm("token")
	title := c.PostForm("title")
	//验证 Token 合法性，提取Id
	Id, err := userService.GetIdFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	//获取 UserInfo
	userDao, err := userService.UserInfoById(Id)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}

	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	filename := filepath.Base(data.Filename)
	finalName := fmt.Sprintf("%s_%s", userDao.Name, filename)
	saveFile := filepath.Join("./public/video/", finalName)
	if err := c.SaveUploadedFile(data, saveFile); err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	videoLocation := "/video/"
	coverLocation := "/cover/" + util.GetFileName(finalName) + ".jpg"
	// 生成封面
	util.GetFrame("./public"+videoLocation+finalName, "./public"+coverLocation)

	videoDao := dao.Video{
		VideoName:     finalName,
		VideoLocation: videoLocation,
		CoverLocation: coverLocation,
		UploaderId:    userDao.Id,
		Title:         title,
	}
	videoService.AddNewVideo(&videoDao)
	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  finalName + " uploaded successfully",
	})
}

// PublishList all users have same publish video list
func PublishList(c *gin.Context) {
	userId, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "empty token"})
		return
	}
	//验证 Token 合法性，提取Id
	_, err := userService.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	userDao, err := userService.UserInfoByUserId(userId)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}
	videoDaoList, err := videoService.GetVideoDaoListById(userDao.Id)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	var videoList []Video
	for _, videoDao := range videoDaoList {

		var name = userDao.Name
		var author = User{
			Id:            userId,
			Name:          name,
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
