package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
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

	videoID, err := global.GVAR_ID_GENERATOR.NextID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	name := strconv.FormatUint(videoID, 10)
	videoName := name + ".mp4"
	coverName := name + ".jpg"

	videoSavePath := filepath.Join(global.GVAR_VIDEO_ADDR, videoName)
	coverSavePath := filepath.Join(global.GVAR_COVER_ADDR, coverName)

	if err = c.SaveUploadedFile(data, videoSavePath); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	if err = util.GetFrame(videoSavePath, coverSavePath); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	err = service.PublishVideo(userID, videoID, videoName, coverName, title)
	if err != nil {
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
	userID := c.GetUint64("UserID")

	var videos []dao.Video
	if err := service.GetPublishedVideos(&videos, userID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	var videoList []Video
	for _, video_ := range videos {
		VideoLocation := global.GVAR_VIDEO_ADDR + video_.PlayName
		if _, err := os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := global.GVAR_VIDEO_ADDR + video_.CoverName
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
			PlayUrl:       "http://" + c.Request.Host + "/static/video/" + video_.PlayName,
			CoverUrl:      "http://" + c.Request.Host + "/static/cover/" + video_.CoverName,
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
