package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/service"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed video list for every request
func Feed(c *gin.Context) {
	println("Feed\n\n\n\n")
	// 不传latest_time默认为当前时间
	var CurrentTimeInt int64 = time.Now().UnixMilli()
	var CurrentTime string = strconv.FormatInt(CurrentTimeInt, 10)
	var LatestTimeStr string = c.DefaultQuery("latest_time", CurrentTime)
	LatestTime, err := strconv.ParseInt(LatestTimeStr, 10, 64)

	if err != nil {
		//无法解析latest_time
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "parameter latest_time is wrong"})
		return
	}
	var videoList []dao.Video
	var authorList []dao.User
	numVideos, err := service.GetFeedVideosAndAuthors(&videoList, &authorList, LatestTime, global.GVAR_FEED_NUM)

	println(numVideos, "@@@\n\n\n")
	if err != nil {
		//访问数据库出错
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	} else if numVideos == 0 {
		//没有满足条件的视频
		c.JSON(http.StatusOK, FeedResponse{
			Response:  Response{StatusCode: 0},
			VideoList: nil,
			NextTime:  CurrentTimeInt,
		})
		return
	}

	var (
		videoJsonList  []Video
		videoJson      Video
		video          dao.Video
		author         dao.User
		authorJson     User
		isFavoriteList []bool
		isFollowList   []bool
		isLogged       = false // 用户是否传入了合法有效的token（是否登录）
		idx            int
	)

	var userID uint64
	// 判断传入的token是否合法，用户是否存在
	if token := c.Query("token"); token != "" {
		claims, err := util.ParseToken(token)
		if err == nil {
			userID = claims.UserID
			if service.IsUserIDExist(userID) {
				isLogged = true
			}
		}
	}

	if isLogged {
		// 当用户登录时 一次性获取用户是否点赞了列表中的视频以及是否关注了视频的作者
		videoIdList := make([]uint64, numVideos)
		authorIdList := make([]uint64, numVideos)
		for idx, video = range videoList {
			videoIdList[idx] = video.VideoID
			authorIdList[idx] = video.AuthorID
		}

		isFavoriteList, err = service.GetFavoriteStatusList(userID, videoIdList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		isFollowList, err = service.GetIsFollowStatusList(userID, authorIdList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}

	}

	var isFavorite bool
	var isFollow bool

	for idx, video = range videoList {
		// 未登录时默认为未关注未点赞
		if isLogged {
			// 当用户登录时，判断是否关注当前作者
			isFollow = isFollowList[idx]
			isFavorite = isFavoriteList[idx]
		} else {
			isFavorite = false
			isFollow = false
		}

		// 二次确认返回的视频与封面是服务器存在的
		VideoLocation := filepath.Join(global.GVAR_VIDEO_ADDR, video.PlayName)
		if _, err = os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := filepath.Join(global.GVAR_COVER_ADDR, video.CoverName)
		if _, err = os.Stat(CoverLocation); err != nil {
			continue
		}

		author = authorList[idx]
		authorJson.ID = author.UserID
		authorJson.Name = author.Name
		authorJson.FollowCount = author.FollowCount
		authorJson.FollowerCount = author.FollowerCount
		authorJson.IsFollow = isFollow

		videoJson.ID = video.VideoID
		videoJson.Author = authorJson
		videoJson.PlayUrl = "http://" + c.Request.Host + "/static/video/" + video.PlayName
		videoJson.CoverUrl = "http://" + c.Request.Host + "/static/cover/" + video.CoverName
		videoJson.FavoriteCount = video.FavoriteCount
		videoJson.CommentCount = video.CommentCount
		videoJson.Title = video.Title
		videoJson.IsFavorite = isFavorite

		videoJsonList = append(videoJsonList, videoJson)
	}

	//本次返回的视频中发布最早的时间
	nextTime := videoList[numVideos-1].CreatedAt.UnixMilli()
	fmt.Println(time.UnixMilli(LatestTime), videoList[numVideos-1].CreatedAt, videoJsonList)

	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: videoJsonList,
		NextTime:  nextTime,
	})
}
