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
		if err := tx.Create(&comment).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}

		if err := tx.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
			return err
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
	if err := global.GVAR_DB.Where("comment_id = ?", commentID).Find(&comment).Error; err != nil {
		return err
	} else if comment.UserID != userID {
		// 删除其他用户评论
		return errors.New("不用删除其他用户评论")
	}

	err := global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		if err := tx.Delete(&comment).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}

		if err := tx.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count -1")).Error; err != nil {
			return err
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
