package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
)

// AddComment 用户userID向视频videoID发送评论，评论内容为commentText
func AddComment(userID uint64, videoID uint64, commentText string) (dao.Comment, error) {
	commentID, err := global.GVAR_ID_GENERATOR.NextID()
	if err != nil {
		return dao.Comment{}, err
	}
	comment := dao.Comment{
		CommentID: commentID,
		VideoID:   videoID,
		UserID:    userID,
		Content:   commentText,
	}

	if err = global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		if result := tx.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count + 1")); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return errors.New("video_id 不存在")
		}
		if result := tx.Create(&comment); result.Error != nil {
			// 返回任何错误都会回滚事务
			return result.Error
		}
		// 返回 nil 提交事务
		return nil
	}); err != nil {
		return dao.Comment{}, err
	}

	return comment, nil
}

// DeleteComment 用户userID删除视频videoID的评论commentID
func DeleteComment(userID uint64, videoID uint64, commentID uint64) error {
	var comment dao.Comment
	comment.CommentID = commentID
	err := global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		if result := tx.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&comment); result.Error != nil {
			// 返回任何错误都会回滚事务
			return result.Error
		} else if result.RowsAffected == 0 {
			return errors.New("Comment 表中 user_id 或 video_id 不存在")
		}

		if result := tx.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count -1")); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return errors.New("video 表中 video_id 不存在")
		}

		// 返回 nil 提交事务
		return nil
	})
	return err
}

// GetCommentListAndUserList 获取评论列表和对应的用户列表
func GetCommentListAndUserList(videoID uint64, commentList *[]dao.Comment, userList *[]dao.User) error {
	result := global.GVAR_DB.Debug().Model(dao.Comment{}).Where("video_id = ?", videoID).Find(commentList)
	if result.Error != nil {
		return result.Error
	}

	userIDList := make([]uint64, 0, result.RowsAffected)
	for _, comment := range *commentList {
		userIDList = append(userIDList, comment.UserID)
	}

	return GetUserListByUserIDs(userIDList, userList)

}
