package redis

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"strconv"
)

func FavoriteAction(videoID, userID, authorID uint64) error {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	videoFavoritedRedis := fmt.Sprintf(VideoFavoritedPattern, videoID)
	userRedis := fmt.Sprintf(UserPattern, userID)
	authorRedis := fmt.Sprintf(UserPattern, authorID)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(CTX, func(pipe redis.Pipeliner) error {
			//查询是否点赞
			if isFavorite := tx.SIsMember(CTX, userFavoriteRedis, videoID).Val(); isFavorite {
				return nil
			}

			// 增加点赞关系
			pipe.SAdd(CTX, userFavoriteRedis, videoID)
			pipe.SAdd(CTX, videoFavoritedRedis, userID)

			//更新redis video相关
			/* Add your code here*/

			//更新redis user相关
			if result := pipe.Exists(CTX, userRedis).Val(); result <= 0 {
				return nil
			}
			pipe.HIncrBy(CTX, userRedis, "favorite_count", 1)
			pipe.HIncrBy(CTX, authorRedis, "total_favorite", 1)

			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, userFavoriteRedis, videoFavoritedRedis, userRedis, authorRedis)
}

func CancelFavorite(videoID, userID, authorID uint64) error {
	//定义 key
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	videoFavoritedRedis := fmt.Sprintf(VideoFavoritedPattern, videoID)
	userRedis := fmt.Sprintf(UserPattern, userID)
	authorRedis := fmt.Sprintf(UserPattern, authorID)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(CTX, func(pipe redis.Pipeliner) error {
			//查询是否点赞
			if isFavorite := tx.SIsMember(CTX, userFavoriteRedis, videoID).Val(); !isFavorite {
				return nil
			}

			// 删除点赞关系
			pipe.SRem(CTX, userFavoriteRedis, videoID)
			pipe.SRem(CTX, videoFavoritedRedis, userID)

			//更新redis video相关
			/* Add your code here*/

			//更新redis user相关
			if result := pipe.Exists(CTX, userRedis).Val(); result <= 0 {
				return nil
			}
			pipe.HIncrBy(CTX, userRedis, "favorite_count", -1)
			pipe.HIncrBy(CTX, authorRedis, "total_favorite", -1)

			return nil
		})
		return err
	}

	// 多次尝试提交
	return Retry(txf, userFavoriteRedis, videoFavoritedRedis, userRedis, authorRedis)
}

// GetFavoriteListByUserID 获取用户点赞列表
func GetFavoriteListByUserID(userID uint64) ([]dao.Video, error) {
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	if result := global.GVAR_REDIS.Exists(CTX, userFavoriteRedis).Val(); result <= 0 {
		return nil, errors.New("Not found in cache")
	}
	videoIDStrList := global.GVAR_REDIS.SMembers(CTX, userFavoriteRedis).Val()
	videoIDList := make([]uint64, 0, len(videoIDStrList))
	for i := 0; i < len(videoIDStrList); i++ {
		videoID, err := strconv.ParseUint(videoIDStrList[i], 10, 64)
		if err != nil {
			return nil, errors.New("Wrong format conversion in cache")
		}
		videoIDList = append(videoIDList, videoID)
	}
	var videoDaoList []dao.Video
	err := GetVideoListByIDs(&videoDaoList, videoIDList)
	if err != nil {
		return nil, err
	}
	return videoDaoList, nil
}

func SetFavoriteListWithUserID(userID uint64, videoDaoList []dao.Video) error {
	videoNum := len(videoDaoList)
	userFavoriteRedis := fmt.Sprintf(UserFavoritePattern, userID)
	userRedis := fmt.Sprintf(UserPattern, userID)
	videoFavoritedRedisList := make([]string, 0, videoNum)
	authorRedisList := make([]string, 0, videoNum)
	for _, videoDao := range videoDaoList {
		videoFavoritedRedis := fmt.Sprintf(VideoFavoritedPattern, videoDao.VideoID)
		videoFavoritedRedisList = append(videoFavoritedRedisList, videoFavoritedRedis)
		authorRedis := fmt.Sprintf(UserPattern, videoDao.AuthorID)
		authorRedisList = append(videoFavoritedRedisList, authorRedis)
	}

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(CTX, func(pipe redis.Pipeliner) error {
			// 增加点赞关系
			for idx, videoDao := range videoDaoList {
				pipe.SAdd(CTX, userFavoriteRedis, videoDao.VideoID)
				pipe.SAdd(CTX, videoFavoritedRedisList[idx], userID)

				//更新redis video相关
				/* Add your code here*/

				//更新redis user相关
				if result := pipe.Exists(CTX, userRedis).Val(); result <= 0 {
					return nil
				}
				pipe.HIncrBy(CTX, userRedis, "favorite_count", 1)
				pipe.HIncrBy(CTX, authorRedisList[idx], "total_favorite", 1)
			}

			return nil
		})
		return err
	}

	// 多次尝试提交
	watchedKeys := make([]string, 0, 2*len(videoFavoritedRedisList)+2)
	watchedKeys = append(watchedKeys, userFavoriteRedis)
	watchedKeys = append(watchedKeys, userRedis)
	watchedKeys = append(watchedKeys, videoFavoritedRedisList...)
	watchedKeys = append(watchedKeys, authorRedisList...)
	return Retry(txf, watchedKeys...)
}
