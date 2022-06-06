package service

import (
	"errors"
	"fmt"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func AddComment(comment *model.Comment) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
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
	var comment model.Comment
	comment.CommentID = commentID
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&comment).Error; err != nil {
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
func GetCommentListAndUserListRedis(videoID uint64, commentList *[]model.Comment, userList *[]model.User) error {
	keyCommentsOfVideo := fmt.Sprintf(VideoCommentsPattern, videoID)
	n, err := global.REDIS.Exists(global.CONTEXT, keyCommentsOfVideo).Result()
	if err != nil {
		return err
	}
	if n <= 0 {
		//	CommentsOfVideo:id 不存在
		// 先去 keyVideo 中check comment_count是否为0
		numComments, err := GetCommentCountOfVideo(videoID)
		if err != nil {
			return err
		}
		if numComments == 0 {
			return nil
		}
		// 不止一条comment或key不存在的话查表
		result := global.DB.Where("video_id = ?", videoID).Find(commentList)
		if result.Error != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return nil
		}
		// 成功
		numComments = int(result.RowsAffected)
		if err = GoCommentsOfVideo(*commentList, keyCommentsOfVideo); err != nil {
			return err
		}
		for _, comment := range *commentList {
			if err = GoComment(comment); err != nil {
				return err
			}
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
	*commentList = make([]model.Comment, 0, numComments)
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
		var comment model.Comment
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

// GetCommentCountListByVideoIDList 被调用当我们不知道videoID是否在redis中
func GetCommentCountListByVideoIDList(videoIDList []uint64, commentCountList *[]int64) error {
	//查询redis
	numVideos := len(videoIDList)
	notInCacheIDList := make([]uint64, 0, numVideos)
	*commentCountList = make([]int64, numVideos)
	inCache := make([]bool, numVideos)
	for i, videoID := range videoIDList {
		keyVideo := fmt.Sprintf(VideoPattern, videoID)
		n, err := global.REDIS.Exists(global.CONTEXT, keyVideo).Result()
		if err != nil {
			return err
		}
		if n <= 0 {
			keyCommentsOfVideo := fmt.Sprintf(VideoCommentsPattern, videoID)
			n, err = global.REDIS.Exists(global.CONTEXT, keyCommentsOfVideo).Result()
			if err != nil {
				return err
			}
			// Video与CommentsOfVideo都不存在
			if n <= 0 {
				notInCacheIDList = append(notInCacheIDList, videoID)
				inCache[i] = false
				continue
			}
			// Video不存在但是CommentsOfVideo存在
			commentCount, err := global.REDIS.ZCard(global.CONTEXT, keyCommentsOfVideo).Uint64()
			if err != nil {
				return err
			}
			(*commentCountList)[i] = int64(commentCount)
			inCache[i] = true
			continue
		}
		// 缓存存在
		commentCount, err := global.REDIS.HGet(global.CONTEXT, keyVideo, "comment_count").Int64()
		if err != nil {
			return err
		}
		(*commentCountList)[i] = commentCount
		inCache[i] = true
	}
	if len(notInCacheIDList) == 0 {
		return nil
	}
	//缓存没有找到，数据库查询
	var commentCountListNotInCache []int64
	if err := GetCommentCountListByVideoIDListSql(notInCacheIDList, &commentCountListNotInCache); err != nil {
		return err
	}
	idxNotInCache := 0
	for i, _ := range *commentCountList {
		if inCache[i] == false {
			(*commentCountList)[i] = commentCountListNotInCache[idxNotInCache]
			idxNotInCache++
		}
	}
	return nil
}

// GetCommentCountListByVideoIDListSql 被调用当且仅当VideoID不在cache中，不得不通过sql查询
func GetCommentCountListByVideoIDListSql(videoIDList []uint64, commentCountList *[]int64) error {
	var uniqueVideoList []model.VideoCount
	result := global.DB.Model(&model.Comment{}).Select("video_id", "COUNT(video_id) as comment_count").
		Where("video_id in ?", videoIDList).Group("video_id").Find(&uniqueVideoList)
	if result.Error != nil {
		return result.Error
	}
	numVideos := result.RowsAffected
	// 针对查询结果建立映射关系
	*commentCountList = make([]int64, 0, numVideos)
	mapVideoIDToCommentCount := make(map[uint64]int64, numVideos)
	for _, each := range uniqueVideoList {
		mapVideoIDToCommentCount[each.VideoID] = each.CommentCount
	}
	for _, videoID := range videoIDList {
		*commentCountList = append(*commentCountList, mapVideoIDToCommentCount[videoID])
	}
	return nil
}
