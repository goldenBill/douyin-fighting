package service

import (
	"errors"
	"fmt"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func AddComment(comment *dao.Comment) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := global.DB.Create(comment).Error; err != nil {
			return err
		}
		//尝试更新 redis 缓存；失败则删除 redis 缓存
		if err := AddCommentInRedis(comment); err != nil {
			return err
		}
		return nil
	})
}

func DeleteComment(userID uint64, videoID uint64, commentID uint64) error {
	var comment dao.Comment
	comment.CommentID = commentID
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := global.DB.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&comment).Error; err != nil {
			return err
		}
		//尝试更新 redis 缓存；失败则删除 redis 缓存
		if err := DeleteCommentInRedis(videoID, commentID); err != nil {
			return err
		}
		return nil
	})
}

// GetCommentListAndUserListRedis 获取评论列表和对应的用户列表
func GetCommentListAndUserListRedis(videoID uint64, commentList *[]dao.Comment, userList *[]dao.User) error {
	keyCommentsOfVideo := fmt.Sprintf(VideoCommentsPattern, videoID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyCommentsOfVideo).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		//	CommentsOfVideo:id 不存在
		result := global.DB.Where("video_id = ?", videoID).Find(commentList)
		if result.Error != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return nil
		}
		// 成功
		numComments := int(result.RowsAffected)
		if err = GoCommentsOfVideo(*commentList, keyCommentsOfVideo); err != nil {
			return err
		}
		authorIDList := make([]uint64, numComments)
		for i, comment := range *commentList {
			authorIDList[i] = comment.UserID
		}
		return GetUserListByUserIDs(authorIDList, userList)
	}
	//	CommentsOfVideo:id 存在
	if err = global.REDIS.Expire(global.CONTEXT, keyCommentsOfVideo, global.VIDEO_EXPIRE).Err(); err != nil {
		return err
	}
	commentIDStrList, err := global.REDIS.ZRevRange(global.CONTEXT, keyCommentsOfVideo, 0, -1).Result()
	if err != nil {
		return err
	}
	numComments := len(commentIDStrList)
	*commentList = make([]dao.Comment, 0, numComments)
	authorIDList := make([]uint64, 0, numComments)

	for _, commentIDStr := range commentIDStrList {
		commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
		keyComment := fmt.Sprintf(CommentPattern, commentID)
		if err != nil {
			continue
		}
		n, err = global.REDIS.Exists(global.CONTEXT, keyComment).Result()
		if err != nil {
			return err
		}
		var comment dao.Comment
		if n <= 0 {
			// "comment_id"不存在
			result := global.DB.Where("comment_id = ?", commentID).Limit(1).Find(&comment)
			if result.Error != nil || result.RowsAffected == 0 {
				return errors.New("get Comment fail")
			}
			if err = GoComment(comment); err != nil {
				continue
			}
			*commentList = append(*commentList, comment)
			authorIDList = append(authorIDList, comment.UserID)
			continue
		}
		if err = global.REDIS.Expire(global.CONTEXT, keyComment, global.VIDEO_EXPIRE).Err(); err != nil {
			continue
		}
		if err = global.REDIS.HGetAll(global.CONTEXT, keyComment).Scan(&comment); err != nil {
			continue
		}
		comment.CommentID = commentID
		timeUnixMilliStr, err := global.REDIS.HGet(global.CONTEXT, keyComment, "created_at").Result()
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
	return GetUserListByUserIDs(authorIDList, userList)
}

func GoComment(comment dao.Comment) error {
	keyComment := "Comment:" + strconv.FormatUint(comment.CommentID, 10)
	err := global.REDIS.HSet(global.CONTEXT, keyComment, "video_id", comment.VideoID,
		"user_id", comment.UserID, "content", comment.Content, "created_at", comment.CreatedAt.UnixMilli()).Err()
	return err
}
