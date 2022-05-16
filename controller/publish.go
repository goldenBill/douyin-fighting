package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"path/filepath"
)

type VideoListResponse struct {
	Response
	VideoList []Video `json:"video_list"`
}

func GetUserInfoByToken(token string) *dao.UserDao {
	var userDao dao.UserDao
	sqlStr := "select * from user where token = ?"
	if err := dao.DataBase.Get(&userDao, sqlStr, token); err != nil {
		fmt.Println("exec failed, ", err)
		return nil
	}
	return &userDao
}

func AddNewVideo(videoName, videoLocation, coverLocation string, uploaderId int64) {
	sqlStr := "insert into video (video_name, video_location, cover_location, uploader_id) values (?, ?, ?, ?)"
	_, err := dao.DataBase.Exec(sqlStr, videoName, videoLocation, coverLocation, uploaderId)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
}

// Publish check token then save upload file to public directory
func Publish(c *gin.Context) {
	token := c.PostForm("token")
	userDao := GetUserInfoByToken(token)

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
	finalName := fmt.Sprintf("%d_%s", userDao.Id, filename)
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

	AddNewVideo(finalName, videoLocation, coverLocation, userDao.Id)
	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  finalName + " uploaded successfully",
	})
}

func GetVideoDaoListById(userId int64, PublishList *[]dao.VideoDao) {
	sqlStr := "select * from video where uploader_id = ? order by id desc"
	if err := dao.DataBase.Select(PublishList, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
}

// PublishList all users have same publish video list
func PublishList(c *gin.Context) {
	token := c.Query("token")
	userDao := GetUserInfoByToken(token)

	var videoDaoList []dao.VideoDao
	GetVideoDaoListById(userDao.Id, &videoDaoList)

	var videoList []Video
	for _, videoDao := range videoDaoList {

		var followCount = GetFollowCount(userDao.Id)
		var followerCount = GetFollowerCount(userDao.Id)
		var name = userDao.Name
		var isFollow = false
		var author = User{
			Id:            videoDao.UploaderId,
			Name:          name,
			FollowCount:   followCount,
			FollowerCount: followerCount,
			IsFollow:      isFollow,
		}
		var favoriteCount = GetFavoriteCount(userDao.Id)
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
