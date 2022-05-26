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
	println("Publish\n\n\n\n")
	// 获取 userID
	userID := c.GetUint64("UserID")
	// 判断userID是否存在
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
		// 无法生成ID
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	name := strconv.FormatUint(videoID, 10)
	videoName := name + c.GetString("FileType")
	coverName := name + ".jpg"

	videoSavePath := filepath.Join(global.GVAR_VIDEO_ADDR, videoName)
	coverSavePath := filepath.Join(global.GVAR_COVER_ADDR, coverName)

	if err = c.SaveUploadedFile(data, videoSavePath); err != nil {
		// 视频无法保存
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	if err = util.GetFrame(videoSavePath, coverSavePath); err != nil {
		// 封面无法保存
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	err = service.PublishVideo(userID, videoID, videoName, coverName, title)
	if err != nil {
		// 无法写库
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
	println("PublishList\n\n\n\n")
	var err error
	// 获取 authorID
	authorID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	var videoList []dao.Video
	var authorList []dao.User
	numVideos, err := service.GetPublishedVideosAndAuthors(&videoList, &authorList, authorID)
	if err != nil {
		//访问数据库出错
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	} else if numVideos == 0 {
		//没有满足条件的视频
		c.JSON(http.StatusOK, FeedResponse{
			Response:  Response{StatusCode: 0},
			VideoList: nil,
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
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: videoJsonList,
	})
}
