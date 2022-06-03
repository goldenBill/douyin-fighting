package service

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
	"time"
)

// AddCommentRedis 用户userID向视频videoID发送评论，评论内容为commentText
func AddCommentRedis(comment *dao.Comment) error {
	if err := global.DB.Create(comment).Error; err != nil {
		return err
	}
	// 写redis
	pipe := global.REDIS.TxPipeline()
	videoIDStr := strconv.FormatUint(comment.VideoID, 10)
	keyCommentsOfVideo := "CommentsOfVideo:" + videoIDStr
	// 确保keyCommentsOfVideo在redis中
	n, err := global.REDIS.Exists(global.CONTEXT, keyCommentsOfVideo).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// KeyCommentsOfVideo cache miss
		var commentList []dao.Comment
		result := global.DB.Where("video_id = ?", comment.VideoID).Find(&commentList)
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("AddComment fail")
		}
		if err = GoCommentsOfVideo(commentList, keyCommentsOfVideo); err != nil {
			return err
		}
	} else {
		// KeyCommentsOfVideo cache hit
		// 直接添加
		Z := redis.Z{Score: float64(comment.CreatedAt.UnixMilli()) / 1000, Member: comment.CommentID}
		pipe.ZAdd(global.CONTEXT, keyCommentsOfVideo, &Z)
	}

	// 确保video在redis中
	if err = CheckVideo(comment.VideoID); err != nil {
		return err
	}

	commentIDStr := strconv.FormatUint(comment.CommentID, 10)
	userIDStr := strconv.FormatUint(comment.UserID, 10)

	keyComment := "Comment:" + commentIDStr
	keyVideo := "Video:" + videoIDStr

	pipe.HIncrBy(global.CONTEXT, keyVideo, "comment_count", 1)
	pipe.HSet(global.CONTEXT, keyComment, "content", comment.Content, "user_id", userIDStr, "video_id", videoIDStr, "created_at", time.Now().UnixMilli())
	_, err = pipe.Exec(global.CONTEXT)

	return err
}

// DeleteCommentRedis 用户userID删除视频videoID的评论commentID
func DeleteCommentRedis(userID uint64, videoID uint64, commentID uint64) error {
	var comment dao.Comment
	comment.CommentID = commentID
	// 先修改mysql
	if err := global.DB.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&comment).Error; err != nil {
		return err
	}

	// 写redis
	pipe := global.REDIS.TxPipeline()
	videoIDStr := strconv.FormatUint(comment.VideoID, 10)
	keyCommentsOfVideo := "CommentsOfVideo:" + videoIDStr
	commentIDStr := strconv.FormatUint(commentID, 10)
	// 确保keyCommentsOfVideo在redis中
	n, err := global.REDIS.Exists(global.CONTEXT, keyCommentsOfVideo).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		// KeyCommentsOfVideo cache miss
		var commentList []dao.Comment
		result := global.DB.Where("video_id = ?", comment.VideoID).Find(&commentList)
		if result.Error != nil || result.RowsAffected == 0 {
			return errors.New("AddComment fail")
		}
		if err = GoCommentsOfVideo(commentList, keyCommentsOfVideo); err != nil {
			return err
		}
	} else {
		// KeyCommentsOfVideo cache hit
		// 直接删除
		pipe.ZRem(global.CONTEXT, keyCommentsOfVideo, commentIDStr)
	}

	// 确保video在redis中
	if err = CheckVideo(comment.VideoID); err != nil {
		return err
	}

	keyComment := "Comment:" + commentIDStr
	keyVideo := "Video:" + videoIDStr
	// KeyComment 是否存在无所谓，因为要被删掉
	pipe.Del(global.CONTEXT, keyComment)
	pipe.HIncrBy(global.CONTEXT, keyVideo, "comment_count", -1)
	_, err = pipe.Exec(global.CONTEXT)
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
