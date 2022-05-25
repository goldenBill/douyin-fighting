package service

import (
	"fmt"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

// AddComment 用户userID向视频videoID发送评论，评论内容为commentText
func AddComment(userID uint64, videoID uint64, commentText string) {
	comment := dao.Comment{
		VideoID: videoID,
		UserID:  userID,
		Content: commentText,
	}
	global.GVAR_DB.Create(&comment)
}

// DeleteComment 用户userID删除视频videoID的评论commentID
func DeleteComment(userID uint64, videoID uint64, commentID uint64) {
	comment := dao.Comment{
		ID:      commentID,
		VideoID: videoID,
		UserID:  userID,
	}
	global.GVAR_DB.Delete(&comment)
}

// GetCommentListAndUserList 获取评论列表和对应的用户列表
func GetCommentListAndUserList(videoID int64) ([]dao.Comment, []dao.User) {
	commentList := make([]dao.Comment, 0, 20)
	userList := make([]dao.User, 0, 20)
	rows, _ := global.GVAR_DB.Model(dao.Comment{}).Where("video_id = ?", videoID).Order("created_at desc").Rows()

	for rows.Next() {
		var comment dao.Comment
		err := global.GVAR_DB.ScanRows(rows, &comment)
		var userDao dao.User
		sqlStr := "select * from user where user_id = ?"
		err = global.GVAR_SQLX_DB.Get(&userDao, sqlStr, comment.UserID)
		if err != nil {
			fmt.Printf("%#v\n", comment)
			fmt.Printf("%#v\n", err)
			fmt.Printf("%#v\n", userDao)
			continue
		}
		commentList = append(commentList, comment)
		userList = append(userList, userDao)
	}

	return commentList, userList
}
