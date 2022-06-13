package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
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

// Feed 视频流接口（给客户端推送短视频）
func Feed(c *gin.Context) {
	// 不传latest_time默认为当前时间
	var CurrentTimeInt = time.Now().UnixMilli()
	var CurrentTime = strconv.FormatInt(CurrentTimeInt, 10)
	var LatestTimeStr = c.DefaultQuery("latest_time", CurrentTime)
	LatestTime, err := strconv.ParseInt(LatestTimeStr, 10, 64)
	if err != nil {
		// 无法解析latest_time
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "parameter latest_time is wrong"})
		return
	}
	// 得到本次要返回的视频以及其作者
	var videoList []model.Video
	var authorList []model.User
	numVideos, err := service.GetFeedVideosAndAuthorsRedis(&videoList, &authorList, LatestTime, global.FEED_NUM)

	if err != nil {
		// 访问数据库出错
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	if numVideos == 0 {
		// 没有满足条件的视频 使用当前时间再获取一遍
		numVideos, _ = service.GetFeedVideosAndAuthorsRedis(&videoList, &authorList, CurrentTimeInt, global.FEED_NUM)
		if numVideos == 0 {
			// 后端没有视频了
			c.JSON(http.StatusOK, FeedResponse{
				Response:  Response{StatusCode: 0},
				VideoList: nil,
				NextTime:  CurrentTimeInt, // 没有视频可刷时返回当前时间
			})
			return
		}
	}

	var (
		videoJsonList  = make([]Video, 0, numVideos)
		videoJson      Video
		author         model.User
		authorJson     User
		isFavoriteList []bool
		isFollowList   []bool
		isLogged       = false // 用户是否传入了合法有效的token（是否登录）
	)

	var userID uint64
	// 判断传入的token是否合法，用户是否存在
	if token := c.Query("token"); token != "" {
		claims, err := util.ParseToken(token)
		if err == nil {
			// token合法
			userID = claims.UserID
			isLogged = true
		}
	}

	if isLogged {
		// 当用户登录时 批量获取用户是否点赞了列表中的视频以及是否关注了视频的作者
		videoIDList := make([]uint64, numVideos)
		authorIDList := make([]uint64, numVideos)
		for i, video := range videoList {
			videoIDList[i] = video.VideoID
			authorIDList[i] = video.AuthorID
		}
		// 批量获取用户是否用视频点赞
		isFavoriteList, err = service.GetFavoriteStatusList(userID, videoIDList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		// 批量获取用户是否关注作者
		isFollowList, err = service.GetFollowStatusList(userID, authorIDList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
	}

	// 未登录时默认为未关注未点赞
	var isFavorite = false
	var isFollow = false

	for i, video := range videoList {
		if isLogged {
			// 当用户登录时，判断是否关注当前作者
			isFollow = isFollowList[i]
			isFavorite = isFavoriteList[i]
		}

		// 二次确认返回的视频与封面是服务器存在的
		VideoLocation := filepath.Join(global.VIDEO_ADDR, video.PlayName)
		if _, err = os.Stat(VideoLocation); err != nil {
			continue
		}
		CoverLocation := filepath.Join(global.COVER_ADDR, video.CoverName)
		if _, err = os.Stat(CoverLocation); err != nil {
			continue
		}
		// 填充JSON返回值
		author = authorList[i]
		authorJson.ID = author.UserID
		authorJson.Name = author.Name
		authorJson.FollowCount = author.FollowCount
		authorJson.FollowerCount = author.FollowerCount
		authorJson.TotalFavorited = author.TotalFavorited
		authorJson.FavoriteCount = author.FavoriteCount
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

	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: videoJsonList,
		NextTime:  nextTime,
	})
}
