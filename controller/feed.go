package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"net/http"
	"strconv"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

func GetVideoDaoList(videoDaoList *[]dao.Video) {
	sqlStr := "select * from video order by id desc limit 0, 10"
	if err := global.GVAR_SQLX_DB.Select(videoDaoList, sqlStr); err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
}

func GetFollowCount(userId uint64) uint64 {
	var followCount uint64
	sqlStr := "select count(*) from follow where follower_id = ?"
	if err := global.GVAR_SQLX_DB.Get(&followCount, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return 0
	}
	return followCount
}

func GetFollowerCount(userId uint64) uint64 {
	var followCount uint64
	sqlStr := "select count(*) from follow where user_id = ?"
	if err := global.GVAR_SQLX_DB.Get(&followCount, sqlStr, userId); err != nil {
		fmt.Println("exec failed, ", err)
		return 0
	}
	return followCount
}

func GetAuthorName(authorId uint64) string {
	var authorName string
	sqlStr := "select name from user where id = ?"
	if err := global.GVAR_SQLX_DB.Get(&authorName, sqlStr, authorId); err != nil {
		fmt.Println("exec failed, ", err)
		return "0"
	}
	return authorName
}

func GetSkipID(authorId uint64) string {
	var authorSkipID string
	sqlStr := "select user_id from user where id = ?"
	if err := global.GVAR_SQLX_DB.Get(&authorSkipID, sqlStr, authorId); err != nil {
		fmt.Println("exec failed, ", err)
		return "0"
	}
	return authorSkipID
}

func GetFavoriteCount(videoId uint64) uint64 {
	var favoriteCount uint64
	sqlStr := "select count(*) from favorite where video_id = ?"
	if err := global.GVAR_SQLX_DB.Get(&favoriteCount, sqlStr, videoId); err != nil {
		fmt.Println("exec failed, ", err)
		return 0
	}
	return favoriteCount
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	latestTimeUnix, _ := strconv.ParseInt(c.DefaultQuery("latest_time", "1652480260 * 1000"), 10, 64)
	//var token = c.DefaultQuery("token", "")

	var videoDaoList []dao.Video
	GetVideoDaoList(&videoDaoList)

	var videoList []Video
	for _, videoDao := range videoDaoList {
		//if videoDao.CreateTime.UnixNano()/1e6 < latestTimeUnix {
		//	fmt.Println(videoDao.CreateTime.Unix(), latestTimeUnix)
		//	continue
		//}

		var followCount = GetFollowCount(videoDao.UploaderId)
		var followerCount = GetFollowerCount(videoDao.UploaderId)
		var name = GetAuthorName(videoDao.UploaderId)
		var authorSkipID, _ = strconv.ParseUint(GetSkipID(videoDao.UploaderId), 10, 64)
		var isFollow = false
		var author = User{
			Id:            authorSkipID,
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
			Title:         *videoDao.Title,
		}
		videoList = append(videoList, video)
	}

	nextTime := latestTimeUnix
	if len(videoDaoList) > 0 {
		nextTime = videoDaoList[len(videoDaoList)-1].CreateTime.UnixNano() / 1e6
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "OK"},
		VideoList: videoList,
		NextTime:  nextTime,
	})
}
