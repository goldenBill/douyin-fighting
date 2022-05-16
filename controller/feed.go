package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"net/http"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

func GetVideoDaoList(videoDaoList *[]dao.VideoDao) {
	sqlStr := "select * from video order by id desc limit 0, 10"
	if err := dao.DataBase.Select(videoDaoList, sqlStr); err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
}

func GetFollowCount(userId int64) int64 {
	var followCount int64
	sqlStr := "select count(*) from follow where follower_id = ?"
	if err := dao.DataBase.Get(&followCount, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return -1
	}
	return followCount
}

func GetFollowerCount(userId int64) int64 {
	var followCount int64
	sqlStr := "select count(*) from follow where user_id = ?"
	if err := dao.DataBase.Get(&followCount, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return -1
	}
	return followCount
}

func GetAuthorName(authorId int64) string {
	var authorName string
	sqlStr := "select name from user where id = ?"
	if err := dao.DataBase.Get(&authorName, sqlStr, authorId); err != nil {
		fmt.Println("exec failed, ", err)
		return ""
	}
	return authorName
}

func GetFavoriteCount(videoId int64) int64 {
	var favoriteCount int64
	sqlStr := "select count(*) from favorite where video_id = ?"
	if err := dao.DataBase.Get(&favoriteCount, sqlStr, videoId); err != nil {
		fmt.Println("exec failed, ", err)
		return -1
	}
	return favoriteCount
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	//var latestTimeUnix, _ = strconv.ParseInt(c.DefaultQuery("latest_time", "1652480260 * 1000"), 10, 64)
	var latestTimeUnix int64 = 1652480260 * 1000
	//var token = c.DefaultQuery("token", "")

	var videoDaoList []dao.VideoDao
	GetVideoDaoList(&videoDaoList)

	var videoList []Video
	for _, videoDao := range videoDaoList {
		if videoDao.CreateTime.UnixNano()/1e6 < latestTimeUnix {
			fmt.Println(videoDao.CreateTime.Unix(), latestTimeUnix)
			continue
		}

		var followCount = GetFollowCount(videoDao.UploaderId)
		var followerCount = GetFollowerCount(videoDao.UploaderId)
		var name = GetAuthorName(videoDao.UploaderId)
		var isFollow = false
		var author = User{
			Id:            videoDao.UploaderId,
			Name:          name,
			FollowCount:   followCount,
			FollowerCount: followerCount,
			IsFollow:      isFollow,
		}
		var favoriteCount = GetFavoriteCount(videoDao.Id)
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

	var nextTime = videoDaoList[len(videoDaoList)-1].CreateTime.UnixNano() / 1e6
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "OK"},
		VideoList: videoList,
		NextTime:  nextTime,
	})
}
