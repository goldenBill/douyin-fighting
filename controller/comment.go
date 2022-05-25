package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/service"
	"net/http"
)

// CommentActionRequest 评论操作的请求
type CommentActionRequest struct {
	UserID      uint64 `form:"user_id" json:"user_id"`
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
func CommentAction(c *gin.Context) {
	// 参数绑定
	var r CommentActionRequest
	err := c.ShouldBind(&r)
	if err != nil {
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

	// 判断 video_id 是否正确

	// 评论操作
	if r.ActionType == 1 {
		commentDao, err := service.AddComment(r.UserID, r.VideoID, r.CommentText)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "comment failed"})
			return
		}
		userDao, _ := service.UserInfoByUserID(commentDao.UserID)

		c.JSON(http.StatusOK, CommentActionResponse{
			Response: Response{StatusCode: 0},
			Comment: Comment{
				ID: commentDao.CommentID,
				User: User{
					ID:            userDao.UserID,
					Name:          userDao.Name,
					FollowCount:   userDao.FollowCount,
					FollowerCount: userDao.FollowerCount,
					IsFollow:      false,
				},
				Content:    commentDao.Content,
				CreateDate: commentDao.CreatedAt.Format("01-02"),
			},
		})
	} else {
		err = service.DeleteComment(r.UserID, r.VideoID, r.CommentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{StatusCode: 1, StatusMsg: "comment failed"})
			return
		}
		c.JSON(http.StatusOK, Response{StatusCode: 0})
	}
}

// CommentList all videos have same demo comment list
func CommentList(c *gin.Context) {
	// 参数绑定
	var r CommentListRequest
	err := c.ShouldBind(&r)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "bind error"})
		return
	}

	commentDaoList, userDaoList := service.GetCommentListAndUserList(r.VideoID)

	commentList := make([]Comment, 0, len(commentDaoList))
	for i := 0; i < len(commentDaoList); i++ {
		user := User{
			ID:            userDaoList[i].ID,
			Name:          userDaoList[i].Name,
			FollowCount:   userDaoList[i].FollowCount,
			FollowerCount: userDaoList[i].FollowerCount,
		}
		comment := Comment{
			ID:         commentDaoList[i].ID,
			User:       user,
			Content:    commentDaoList[i].Content,
			CreateDate: commentDaoList[i].CreatedAt.Format("01-02"),
		}
		commentList = append(commentList, comment)
	}
	c.JSON(http.StatusOK, CommentListResponse{
		Response:    Response{StatusCode: 0},
		CommentList: commentList,
	})
}
