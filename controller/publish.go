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
	"unicode/utf8"
)

type VideoListResponse struct {
	Response
	VideoList []Video `json:"video_list"`
}

// Publish 投稿接口
func Publish(c *gin.Context) {
	// 获取 userID
	userID := c.GetUint64("UserID")

	title := c.PostForm("title")
	// 判断title是否合法
	if utf8.RuneCountInString(title) > global.MAX_TITLE_LENGTH ||
		utf8.RuneCountInString(title) <= 0 {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "非法视频描述"})
		return
	}

	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	// 生成唯一ID
	videoID, err := global.ID_GENERATOR.NextID()
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

	videoSavePath := filepath.Join(global.VIDEO_ADDR, videoName)
	coverSavePath := filepath.Join(global.COVER_ADDR, coverName)

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

	// 写入数据库
	err = service.PublishVideo(userID, videoID, videoName, coverName, title)

	if err != nil {
		// 无法写入数据库
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

// PublishList 发布列表接口
func PublishList(c *gin.Context) {
	// 获取 authorID
	authorID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  "user_id不合法",
		})
		return
	}
	// 得到用户发布过的视频
	var videoList []model.Video
	numVideos, err := service.GetPublishedVideosRedis(&videoList, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	// 作者相同，无需重复查询
	var author *model.User
	author, err = service.UserInfoByUserID(authorID)
	if err != nil {
		//访问数据库出错
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}
	if numVideos == 0 {
		//没有满足条件的视频
		c.JSON(http.StatusOK, VideoListResponse{
			Response:  Response{StatusCode: 0},
			VideoList: nil,
		})
		return
	}

	var (
		videoJsonList  []Video
		videoJson      Video
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

		isFavoriteList, err = service.GetFavoriteStatusList(userID, videoIDList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
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
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: videoJsonList,
	})
}
