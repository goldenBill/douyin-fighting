package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"github.com/goldenBill/douyin-fighting/util"
)

// Register 用户注册
func Register(username string, password string) (user *model.User, err error) {
	//判断用户名是否存在
	result := global.DB.Where("name = ?", username).Limit(1).Find(&user)
	if result.RowsAffected != 0 {
		err = errors.New("user already exists")
		return
	}
	user.Name = username                          //接收姓名
	user.Password = util.BcryptHash(password)     //对明文密码加密
	user.UserID, _ = global.ID_GENERATOR.NextID() //生成增长的 userID
	err = global.DB.Create(user).Error            //存储到数据库
	return
}

// Login 用户登录
func Login(username string, password string) (user *model.User, err error) {
	//检查用户名是否存在
	result := global.DB.Where("name = ?", username).Limit(1).Find(&user)
	if result.RowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	//检查密码是否正确
	if ok := util.BcryptCheck(password, user.Password); !ok {
		err = errors.New("wrong password")
		return
	}
	return
}

// UserInfoByUserID 通过 UserID 获取用户信息
func UserInfoByUserID(userID uint64) (user *model.User, err error) {
	// 查询缓存
	user, err = GetUserInfoByUserIDFromRedis(userID)
	if err == nil {
		return
	} else if err.Error() != "not found in cache" {
		return nil, err
	}
	// 检查 userID 是否存在；若存在，获取用户信息
	result := global.DB.Where("user_id = ?", userID).Limit(1).Find(&user)
	// 查询关注数目
	global.DB.Model(&model.Follow{}).Where("follower_id = ? and is_follow = ?", user.UserID, true).
		Count(&user.FollowCount)
	// 查询粉丝数目
	global.DB.Model(&model.Follow{}).Where("celebrity_id = ? and is_follow = ?", user.UserID, true).
		Count(&user.FollowerCount)
	// 查询点赞数目
	global.DB.Model(&model.Favorite{}).Where("user_id = ? and is_favorite = ?", user.UserID, true).
		Count(&user.FavoriteCount)
	// 查询总点赞数目
	var publishVideoIDList []uint64
	_ = GetVideoIDListByUserID(user.UserID, &publishVideoIDList)
	favoriteCountList, _ := GetFavoriteCountListByVideoIDList(publishVideoIDList)
	var totalFavorited int64 = 0
	for _, each := range favoriteCountList {
		totalFavorited += each
	}
	user.TotalFavorited = totalFavorited
	if result.RowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	// 更新缓存
	if err = AddUserInfoByUserIDFromCacheToRedis(user); err != nil {
		return nil, err
	}
	return

}

// GetUserListByUserIDList 根据 UserIDList 获取对应的用户列表
func GetUserListByUserIDList(UserIDList []uint64) ([]model.User, error) {
	// 查询缓存
	userList, notInCache, err := GetUserListByUserIDListFromRedis(UserIDList)
	if err != nil && err.Error() != "not found in cache" {
		return nil, err
	} else if err == nil {
		return userList, nil
	}
	var uniqueUserList []model.User
	result := global.DB.Where("user_id in ?", notInCache).Find(&uniqueUserList)
	if result.Error != nil {
		return nil, result.Error
	}
	// 针对查询结果建立映射关系
	mapUserIDToUser := make(map[uint64]model.User, len(uniqueUserList))
	for idx, user := range uniqueUserList {
		// 查询关注数目
		global.DB.Model(&model.Follow{}).Where("follower_id = ? and is_follow = ?", user.UserID, true).
			Count(&uniqueUserList[idx].FollowCount)
		// 查询粉丝数目
		global.DB.Model(&model.Follow{}).Where("celebrity_id = ? and is_follow = ?", user.UserID, true).
			Count(&uniqueUserList[idx].FollowerCount)
		// 查询点赞数目
		global.DB.Model(&model.Favorite{}).Where("user_id = ? and is_favorite = ?", user.UserID, true).
			Count(&uniqueUserList[idx].FavoriteCount)
		// 查询总点赞数目
		var publishVideoIDList []uint64
		_ = GetVideoIDListByUserID(user.UserID, &publishVideoIDList)
		favoriteCountList, _ := GetFavoriteCountListByVideoIDList(publishVideoIDList)
		var totalFavorited int64 = 0
		for _, each := range favoriteCountList {
			totalFavorited += each
		}
		uniqueUserList[idx].TotalFavorited = totalFavorited
		mapUserIDToUser[user.UserID] = uniqueUserList[idx]
	}
	// 更新缓存
	if err = AddUserListByUserIDListsToRedis(uniqueUserList); err != nil {
		return nil, err
	}
	// 后续操作，返回用户列表
	for idx, each := range userList {
		if user, ok := mapUserIDToUser[each.UserID]; ok {
			userList[idx] = user
		}
	}
	return userList, nil
}

// GetUserListByUserIDs 根据UserIDs获取对应的用户列表
func GetUserListByUserIDs(UserIDs []uint64, userList *[]model.User) (err error) {
	userListPrototype, err := GetUserListByUserIDList(UserIDs)
	*userList = userListPrototype
	return
}
