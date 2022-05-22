package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
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

// Publish : check token then save upload file to public directory
func Publish(c *gin.Context) {
	//验证 Token 合法性，提取ID
	tokenString := c.PostForm("token")
	ID, err := userService.GetIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	//获取 UserInfo
	userDao, err := userService.UserInfoByID(ID)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}

	title := c.PostForm("title")
	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	filename := filepath.Base(data.Filename)
	finalName := fmt.Sprintf("%s_%s", userDao.Name, filename)
	videoDao := dao.Video{
		Title:    title,
		PlayUrl:  "/video/" + finalName,
		CoverUrl: "/cover/" + util.GetFileName(finalName) + ".jpg",
		UserID:   userDao.UserID,
		Active:   false,
	}
	// 添加到数据库
	err = videoService.PublishVideo(&videoDao)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	//存储视频
	VideoSavePath := filepath.Join("./public/", videoDao.PlayUrl)
	if err := c.SaveUploadedFile(data, VideoSavePath); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	//生成封面
	CoverSavePath := filepath.Join("./public/", videoDao.CoverUrl)
	if err = util.GetFrame(VideoSavePath, CoverSavePath); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	if err = videoService.SetActive(videoDao.VideoID); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	//成功上传
	c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "uploaded successfully"})
}

// PublishList all users have same publish video list
func PublishList(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "empty token"})
		return
	}
	//验证 Token 合法性，提取ID
	_, err := userService.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	userDao, err := userService.UserInfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}

	videoDaoList, err := videoService.GetPublishedVideos(userID)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	var videoList []Video
	for _, videoDao := range videoDaoList {
		VideoLocation := "./public/" + videoDao.PlayUrl
		if _, err := os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := "./public/" + videoDao.CoverUrl
		if _, err := os.Stat(CoverLocation); err != nil {
			continue
		}
		var author = User{
			ID:            userID,
			Name:          userDao.Name,
			FollowCount:   userDao.FollowCount,
			FollowerCount: userDao.FollowerCount,
			IsFollow:      false,
		}
		video := Video{
			ID:            videoDao.VideoID,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static" + videoDao.PlayUrl,
			CoverUrl:      "http://" + c.Request.Host + "/static" + videoDao.CoverUrl,
			FavoriteCount: videoDao.FavoriteCount,
			CommentCount:  videoDao.CommentCount,
			Title:         videoDao.Title,
			IsFavorite:    false,
		}
		videoList = append(videoList, video)
	}

	c.JSON(http.StatusOK, VideoListResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "OK"},
		VideoList: videoList,
	})
}
