package service

import (
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
)

// AddComment 用户userID向视频videoID发送评论，评论内容为commentText
func AddComment(userID uint64, videoID uint64, commentText string) error {
	commentID, _ := global.GVAR_ID_GENERATOR.NextID()
	comment := dao.Comment{
		CommentID: commentID,
		VideoID:   videoID,
		UserID:    userID,
		Content:   commentText,
	}
	err := global.GVAR_DB.Create(&comment).Error
	if err != nil {
		return err
	}
	return CommentCountPlus(videoID) // 评论+1
}

// DeleteComment 用户userID删除视频videoID的评论commentID
func DeleteComment(userID uint64, videoID uint64, commentID uint64) error {
	comment := dao.Comment{
		ID:      commentID,
		VideoID: videoID,
		UserID:  userID,
	}
	err := global.GVAR_DB.Delete(&comment).Error
	if err != nil {
		return err
	}
	return CommentCountMinus(videoID) // 评论-1
}

// GetCommentListAndUserList 获取评论列表和对应的用户列表
func GetCommentListAndUserList(videoID uint64) ([]dao.Comment, []dao.User) {
	commentList := make([]dao.Comment, 0, 20)
	userIdList := make([]uint64, 0, 20)
	rows, _ := global.GVAR_DB.Model(dao.Comment{}).Where("video_id = ?", videoID).Rows()

	for rows.Next() {
		var comment dao.Comment
		err := global.GVAR_DB.ScanRows(rows, &comment)
		if err != nil {
			continue
		}
		userIdList = append(userIdList, comment.UserID)
		commentList = append(commentList, comment)
	}
	userList, err := GetUserListByUserIDs(userIdList)
	if err != nil {
		return []dao.Comment{}, []dao.User{}
	}

	return commentList, userList
}
