package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"strconv"
	"time"
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

// AddComment 用户userID向视频videoID发送评论，评论内容为commentText
func AddCommentRedis(userID uint64, videoID uint64, commentText string) (dao.Comment, error) {
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

	keyVideo := "Video:" + strconv.FormatUint(videoID, 10)
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyVideo).Result()
	if err != nil {
		return dao.Comment{}, err
	}
	if n <= 0 {
		// "video_id"不存在
		var video dao.Video
		result := global.GVAR_DB.Where("video_id = ?", videoID).Limit(1).Find(&video)
		if result.Error != nil {
			return dao.Comment{}, err
		}
		err = GoVideo(video)
	}

	keyCommentOfVideo := "CommentsOfVideo:" + strconv.FormatUint(videoID, 10)
	keyComment := "Comment:" + strconv.FormatUint(commentID, 10)
	pipe := global.GVAR_REDIS.TxPipeline()
	pipe.HIncrBy(global.GVAR_CONTEXT, keyVideo, "comment_count", 1)
	pipe.LPush(global.GVAR_CONTEXT, keyCommentOfVideo, commentID)
	pipe.HSet(global.GVAR_CONTEXT, keyComment, "content", commentText, "user_id", userID, "video_id", videoID, "created_at", time.Now().UnixMilli())
	_, err = pipe.Exec(global.GVAR_CONTEXT)

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
func DeleteCommentRedis(userID uint64, videoID uint64, commentID uint64) error {
	var comment dao.Comment
	keyComment := "Comment:" + strconv.FormatUint(commentID, 10)
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyComment).Result()
	if err != nil {
		return err
	} else if n > 0 {
		// KeyComment 存在
		if err = global.GVAR_REDIS.HGetAll(global.GVAR_CONTEXT, keyComment).Scan(&comment); err != nil {
			return err
		} else if comment.VideoID == videoID && comment.UserID == userID {
			keyCommentOfVideo := "CommentsOfVideo:" + strconv.FormatUint(videoID, 10)
			keyVideo := "Video:" + strconv.FormatUint(videoID, 10)
			pipe := global.GVAR_REDIS.TxPipeline()
			pipe.Del(global.GVAR_CONTEXT, keyComment)
			pipe.LRem(global.GVAR_CONTEXT, keyCommentOfVideo, 1, commentID)
			pipe.HIncrBy(global.GVAR_CONTEXT, keyVideo, "comment_count", -1)
			_, err = pipe.Exec(global.GVAR_CONTEXT)
		} else {
			// 非法参数
			return errors.New("非法请求")
		}
	}

	comment.CommentID = commentID
	err = global.GVAR_DB.Transaction(func(tx *gorm.DB) error {
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

// GetCommentListAndUserList 获取评论列表和对应的用户列表
func GetCommentListAndUserListRedis(videoID uint64, commentList *[]dao.Comment, userList *[]dao.User) error {
	keyCommentOfVideo := "CommentsOfVideo:" + strconv.FormatUint(videoID, 10)
	n, err := global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, keyCommentOfVideo).Result()
	if err != nil {
		return err
	} else if n <= 0 {
		//	CommentsOfVideo:id 不存在
		result := global.GVAR_DB.Where("video_id = ?", videoID).Find(commentList)
		if result.Error != nil {
			return err
		} else if result.RowsAffected == 0 {
			return nil
		} else {
			// 成功
			var commentIDList = make([]string, len(*commentList))

			for i, comment := range *commentList {
				commentIDList[i] = strconv.FormatUint(comment.CommentID, 10)
			}
			GoCommentsOfVideo(commentIDList, videoID)
			authorIDList := make([]uint64, result.RowsAffected)
			for i, comment := range *commentList {
				authorIDList[i] = comment.UserID
			}
			return GetUserListByUserIDs(authorIDList, userList)
		}
	} else {
		//	CommentsOfVideo:id 存在
		vals, err := global.GVAR_REDIS.LRange(global.GVAR_CONTEXT, keyCommentOfVideo, 0, -1).Result()
		if err != nil {
			return err
		}
		*commentList = make([]dao.Comment, 0, len(vals))
		authorIDList := make([]uint64, 0, len(vals))

		for _, comment_id := range vals {
			var comment dao.Comment
			n, err = global.GVAR_REDIS.Exists(global.GVAR_CONTEXT, "Comment:"+comment_id).Result()
			if err != nil {
				return err
			}
			if n <= 0 {
				// "comment_id"不存在
				result := global.GVAR_DB.Where("comment_id = ?", comment_id).Limit(1).Find(&comment)

				err = GoComment(comment)

				if result.Error != nil || result.RowsAffected == 0 {
					return errors.New("Get Comment fail")
				} else {
					*commentList = append(*commentList, comment)
					authorIDList = append(authorIDList, comment.UserID)
				}
				continue
			}
			if err = global.GVAR_REDIS.HGetAll(global.GVAR_CONTEXT, "Comment:"+comment_id).Scan(&comment); err != nil {
				return err
			} else {
				comment.CommentID, err = strconv.ParseUint(comment_id, 10, 64)
				if err != nil {
					continue
				}
				timeUnixMilliStr, err := global.GVAR_REDIS.HGet(global.GVAR_CONTEXT, "Comment:"+comment_id, "created_at").Result()
				if err != nil {
					continue
				}
				timeUnixMilli, err := strconv.ParseInt(timeUnixMilliStr, 10, 64)
				if err != nil {
					continue
				}
				comment.CreatedAt = time.UnixMilli(timeUnixMilli)

				*commentList = append(*commentList, comment)
				authorIDList = append(authorIDList, comment.UserID)
			}
		}
		return GetUserListByUserIDs(authorIDList, userList)
	}
}

func GoComment(comment dao.Comment) error {
	keyComment := "Comment:" + strconv.FormatUint(comment.CommentID, 10)
	err := global.GVAR_REDIS.HSet(global.GVAR_CONTEXT, keyComment, "video_id", comment.VideoID,
		"user_id", comment.UserID, "content", comment.Content, "created_at", comment.CreatedAt.UnixMilli()).Err()
	return err
}
