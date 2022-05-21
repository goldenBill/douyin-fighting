package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller/service"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"path/filepath"
	"strconv"
)

type VideoListResponse struct {
	Response
	VideoList []Video `json:"video_list"`
}

func AddNewVideo(videoName, videoLocation, coverLocation string, uploaderId uint64, title string) {
	sqlStr := "insert into video (video_name, video_location, cover_location, uploader_id, title) values (?, ?, ?, ?, ?)"
	_, err := global.GVAR_SQLX_DB.Exec(sqlStr, videoName, videoLocation, coverLocation, uploaderId, title)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
}

func GetUserInfoByUserId(userId uint64) *dao.User {
	var userDao dao.User
	sqlStr := "select * from user where user_id = ?"
	if err := global.GVAR_SQLX_DB.Get(&userDao, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return nil
	}
	return &userDao
}

// Publish check token then save upload file to public directory
func Publish(c *gin.Context) {
	tokenString := c.PostForm("token")
	title := c.PostForm("title")
	token, err := userService.ParseToken(tokenString)
	claims := token.Claims.(*service.UserClaims)
	userId := claims.UserID
	userDao := GetUserInfoByUserId(userId)

	if userDao == nil {
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

	AddNewVideo(finalName, videoLocation, coverLocation, userDao.Id, title)
	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  finalName + " uploaded successfully",
	})
}

func GetVideoDaoListById(userId uint64, PublishList *[]dao.Video) {
	sqlStr := "select * from video where uploader_id = ? order by id desc"
	if err := global.GVAR_SQLX_DB.Select(PublishList, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
}

// PublishList all users have same publish video list
func PublishList(c *gin.Context) {
	userId, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "empty token"})
		return
	}

	_, err := userService.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	userDao := GetUserInfoByUserId(userId)
	var videoDaoList []dao.Video
	GetVideoDaoListById(userDao.Id, &videoDaoList)

	var videoList []Video
	for _, videoDao := range videoDaoList {

		var followCount = GetFollowCount(userId)
		var followerCount = GetFollowerCount(userId)
		var name = userDao.Name
		var isFollow = false
		var author = User{
			Id:            userId,
			Name:          name,
			FollowCount:   followCount,
			FollowerCount: followerCount,
			IsFollow:      isFollow,
		}
		var favoriteCount = GetFavoriteCount(userId)
		var isFavorite = false
		video := Video{
			Id:            videoDao.Id,
			Author:        author,
			PlayUrl:       "http://" + c.Request.Host + "/static" + videoDao.VideoLocation + videoDao.VideoName,
			CoverUrl:      "http://" + c.Request.Host + "/static" + videoDao.CoverLocation,
			FavoriteCount: favoriteCount,
			CommentCount:  0,
			IsFavorite:    isFavorite,
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
