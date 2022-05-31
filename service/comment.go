package service

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// AddCommentRedis 用户userID向视频videoID发送评论，评论内容为commentText
func AddCommentRedis(comment *dao.Comment) error {
	// 先写mysql
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 事务
		// 评论数加1
		result := tx.Model(&dao.Video{}).Where("video_id = ?", comment.VideoID).Update("comment_count", gorm.Expr("comment_count + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("video_id 不存在")
		}
		// 创建评论
		return tx.Create(comment).Error
	}); err != nil {
		return err
	}

	commentIDStr := strconv.FormatUint(comment.CommentID, 10)
	videoIDStr := strconv.FormatUint(comment.VideoID, 10)
	userIDStr := strconv.FormatUint(comment.UserID, 10)

	keyCommentsOfVideo := "CommentsOfVideo:" + videoIDStr
	keyComment := "Comment:" + commentIDStr
	keyVideo := "Video:" + videoIDStr

	// 确保 KeyCommentsOfVideo存在
	if err := CheckCommentsOfVideo(comment.VideoID); err != nil {
		return err
	}
	// 确保keyVideo存在
	if err := CheckVideo(comment.VideoID); err != nil {
		return err
	}

	// 再更新缓存
	Z := redis.Z{Score: float64(comment.CreatedAt.UnixMilli()) / 1000, Member: comment.CommentID}
	pipe := global.REDIS.TxPipeline()
	pipe.HIncrBy(global.CONTEXT, keyVideo, "comment_count", 1)
	pipe.ZAdd(global.CONTEXT, keyCommentsOfVideo, &Z)
	pipe.HSet(global.CONTEXT, keyComment, "content", comment.Content, "user_id", userIDStr, "video_id", videoIDStr, "created_at", time.Now().UnixMilli())
	_, err := pipe.Exec(global.CONTEXT)

	return err
}

// DeleteCommentRedis 用户userID删除视频videoID的评论commentID
func DeleteCommentRedis(userID uint64, videoID uint64, commentID uint64) error {
	var comment dao.Comment
	comment.CommentID = commentID
	// 先修改mysql
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		result := tx.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&comment)
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("DeleteComment fail")
		}

		result = tx.Model(&dao.Video{}).Where("video_id = ?", videoID).Update("comment_count", gorm.Expr("comment_count -1"))
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("DeleteComment fail")
		}
		// 返回 nil 提交事务
		return nil
	}); err != nil {
		return err
	}

	commentIDStr := strconv.FormatUint(commentID, 10)
	videoIDStr := strconv.FormatUint(videoID, 10)
	keyComment := "Comment:" + commentIDStr
	keyVideo := "Video:" + videoIDStr
	keyCommentsOfVideo := "CommentsOfVideo:" + videoIDStr
	// 确保 KeyCommentsOfVideo存在
	if err := CheckCommentsOfVideo(videoID); err != nil {
		return err
	}
	// 确保keyVideo存在
	if err := CheckVideo(videoID); err != nil {
		return err
	}
	// KeyComment 是否存在无所谓，因为要被删掉
	pipe := global.REDIS.TxPipeline()
	pipe.Del(global.CONTEXT, keyComment)
	pipe.ZRem(global.CONTEXT, keyCommentsOfVideo, commentIDStr)
	pipe.HIncrBy(global.CONTEXT, keyVideo, "comment_count", -1)
	_, err := pipe.Exec(global.CONTEXT)
	return err
}

// GetCommentListAndUserListRedis 获取评论列表和对应的用户列表
func GetCommentListAndUserListRedis(videoID uint64, commentList *[]dao.Comment, userList *[]dao.User) error {
	keyCommentsOfVideo := "CommentsOfVideo:" + strconv.FormatUint(videoID, 10)
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
		// 翻转 commentList: 最近的评论放在前面
		numComments := int(result.RowsAffected)
		for i, j := 0, numComments-1; i < j; i, j = i+1, j-1 {
			(*commentList)[i], (*commentList)[j] = (*commentList)[j], (*commentList)[i]
		}
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
	commentIDStrList, err := global.REDIS.ZRevRange(global.CONTEXT, keyCommentsOfVideo, 0, -1).Result()
	if err != nil {
		return err
	}
	numComments := len(commentIDStrList)
	*commentList = make([]dao.Comment, 0, numComments)
	authorIDList := make([]uint64, 0, numComments)

	for _, commentIDStr := range commentIDStrList {
		commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
		keyComment := "Comment:" + commentIDStr
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
		if err = global.REDIS.HGetAll(global.CONTEXT, keyComment).Scan(&comment); err != nil {
			return err
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
