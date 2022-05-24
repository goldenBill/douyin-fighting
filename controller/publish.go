package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/service"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type VideoListResponse struct {
	Response
	VideoList []Video `json:"video_list"`
}

// Publish check token then save upload file to public directory
func Publish(c *gin.Context) {
	// 获取 userID
	userID := c.GetUint64("UserID")

	if !service.IsUserIDExist(userID) {
		c.JSON(http.StatusForbidden, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}
	title := c.PostForm("title")

	data, err := c.FormFile("data")
	if err != nil {
		// 状态码不确定
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	videoName, coverName, videoID, err := service.PublishVideo(userID, title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	VideoSavePath := filepath.Join("./public/", videoName)

	if err := c.SaveUploadedFile(data, VideoSavePath); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	CoverSavePath := filepath.Join("./public/", coverName)

	if err = util.GetFrame(VideoSavePath, CoverSavePath); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	if err = service.SetActive(videoID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  " uploaded successfully",
	})
}

// PublishList all users have same publish video list
func PublishList(c *gin.Context) {
	// 获取 userID
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	var videos []dao.Video
	if err := service.GetPublishedVideos(&videos, userID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	var videoList []Video
	for _, video_ := range videos {
		VideoLocation := "./public/" + video_.PlayUrl
		if _, err := os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := "./public/" + video_.CoverUrl
		if _, err := os.Stat(CoverLocation); err != nil {
			continue
		}
		author_, err := service.UserInfoByUserID(video_.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		isFollow := false
		author := User{
			ID:            author_.UserID,
			Name:          author_.Name,
			FollowCount:   author_.FollowerCount,
			FollowerCount: author_.FollowerCount,
			IsFollow:      isFollow,
		}
		isFavorite := false
		video := Video{
			ID:            video_.VideoID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static/" + video_.PlayUrl,
			CoverUrl:      "http://" + c.Request.Host + "/static/" + video_.CoverUrl,
			FavoriteCount: video_.FavoriteCount,
			CommentCount:  video_.CommentCount,
			Title:         video_.Title,
			IsFavorite:    isFavorite,
		}
		videoList = append(videoList, video)
	}
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: videoList,
	})
}
