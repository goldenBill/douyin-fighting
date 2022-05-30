package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/service"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
	"unicode/utf8"
)

// CommentActionRequest 评论操作的请求
type CommentActionRequest struct {
	UserID      uint64 `form:"user_id" json:"user_id"` // apk并没有传user_id这个参数
	Token       string `form:"token" json:"token"`
	VideoID     uint64 `form:"video_id" json:"video_id"`
	ActionType  uint   `form:"action_type" json:"action_type"`
	CommentText string `form:"comment_text" json:"comment_text"`
	CommentID   uint64 `form:"comment_id" json:"comment_id"`
}

type CommentActionResponse struct {
	Response
	Comment Comment `json:"comment,omitempty"`
}

// CommentListRequest 评论列表的请求
type CommentListRequest struct {
	UserID  uint64 `form:"user_id" json:"user_id"`
	Token   string `form:"token" json:"token"`
	VideoID uint64 `form:"video_id" json:"video_id"`
}

// CommentListResponse 评论列表的响应
type CommentListResponse struct {
	Response
	CommentList []Comment `json:"comment_list,omitempty"`
}

// CommentAction no practical effect, just check if token is valid
// 1. 确保操作类型正确 2. 确保video_id正确 3. 确保当前用户有权限删除
func CommentAction(c *gin.Context) {
	// 参数绑定
	var r CommentActionRequest
	if err := c.ShouldBind(&r); err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}

	// 判断 action_type 是否正确
	if r.ActionType != 1 && r.ActionType != 2 {
		// action_type 不合法
		c.JSON(http.StatusBadRequest, Response{StatusCode: 1, StatusMsg: "action type error"})
		return
	}

	// 获取 userID
	r.UserID = c.GetUint64("UserID")

	// 评论操作
	if r.ActionType == 1 {
		// 判断comment是否合法
		if utf8.RuneCountInString(r.CommentText) > global.MAX_COMMENT_LENGTH ||
			utf8.RuneCountInString(r.CommentText) <= 0 {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "非法评论"})
			return
		}
		// 添加评论
		commentID, err := global.ID_GENERATOR.NextID()
		if err != nil {
			// 生成ID失败
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "生成评论ID失败"})
			return
		}
		commentDao := dao.Comment{
			CommentID: commentID,
			VideoID:   r.VideoID,
			UserID:    r.UserID,
			Content:   r.CommentText,
		}
		if err = service.AddCommentRedis(&commentDao); err != nil {
			// 评论失败
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "comment failed"})
			return
		}
		userDao, err := service.UserInfoByUserID(commentDao.UserID)
		if err != nil {
			// 未找到评论的用户
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "comment failed"})
			return
		}
		// 批量判断用户是否关注
		isFollow, err := service.GetIsFollowStatus(r.UserID, userDao.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
		// 返回JSON
		c.JSON(http.StatusOK, CommentActionResponse{
			Response: Response{StatusCode: 0},
			Comment: Comment{
				ID: commentDao.CommentID,
				User: User{
					ID:             userDao.UserID,
					Name:           userDao.Name,
					FollowCount:    userDao.FollowCount,
					FollowerCount:  userDao.FollowerCount,
					TotalFavorited: userDao.TotalFavorited,
					FavoriteCount:  userDao.FavoriteCount,
					IsFollow:       isFollow,
				},
				Content:    commentDao.Content,
				CreateDate: commentDao.CreatedAt.Format("2006-01-02 15:04:05 Monday"),
			},
		})
		return
	}

	// 删除评论
	if err := service.DeleteCommentRedis(r.UserID, r.VideoID, r.CommentID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "comment failed"})
		return
	}
	c.JSON(http.StatusOK, Response{StatusCode: 0})
}

// CommentList all videos have same demo comment list
func CommentList(c *gin.Context) {
	// 参数绑定
	var r CommentListRequest
	if err := c.ShouldBind(&r); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}

	var commentDaoList []dao.Comment
	var userDaoList []dao.User
	// 获取评论列表以及对应的作者
	if err := service.GetCommentListAndUserListRedis(r.VideoID, &commentDaoList, &userDaoList); err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
		return
	}

	var (
		isFollowList []bool
		isLogged     = false // 用户是否传入了合法有效的token（是否登录）
		isFollow     bool
		err          error
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
		// 当用户登录时 一次性获取用户是否点赞了列表中的视频以及是否关注了评论的作者
		authorIDList := make([]uint64, len(commentDaoList))
		for i, user_ := range userDaoList {
			authorIDList[i] = user_.UserID
		}
		// 批量判断用户是否关注评论的作者
		isFollowList, err = service.GetIsFollowStatusList(userID, authorIDList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: err.Error()})
			return
		}
	}

	var (
		commentJsonList = make([]Comment, 0, len(commentDaoList))
		commentJson     Comment
		userJson        User
		user            dao.User
	)

	for i, comment := range commentDaoList {
		// 未登录时默认为未关注未点赞
		isFollow = false
		if isLogged {
			// 当用户登录时，判断是否关注当前作者
			isFollow = isFollowList[i]
		}
		user = userDaoList[i]
		userJson.ID = user.UserID
		userJson.Name = user.Name
		userJson.FollowCount = user.FollowCount
		userJson.FollowerCount = user.FollowerCount
		userJson.TotalFavorited = user.TotalFavorited
		userJson.FavoriteCount = user.FavoriteCount
		userJson.IsFollow = isFollow

		commentJson.ID = comment.CommentID
		commentJson.User = userJson
		commentJson.Content = comment.Content
		commentJson.CreateDate = comment.CreatedAt.Format("2006-01-02 15:04:05 Monday")

		commentJsonList = append(commentJsonList, commentJson)
	}
	c.JSON(http.StatusOK, CommentListResponse{
		Response:    Response{StatusCode: 0},
		CommentList: commentJsonList,
	})
}
