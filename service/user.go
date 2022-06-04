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
	//查询redis
	user, err = GetUserInfoByUserIDFromRedis(userID)
	if err == nil {
		return
	} else if err.Error() != "Not found in cache" {
		return nil, err
	}
	//检查 userID 是否存在；若存在，获取用户信息
	result := global.DB.Where("user_id = ?", userID).Limit(1).Find(&user)
	user.FollowCount, _ = GetFollowCountByUserID(user.UserID)
	user.FollowerCount, _ = GetFollowerCountByUserID(user.UserID)
	user.FavoriteCount, _ = GetFavoriteCountByUserID(user.UserID)
	user.TotalFavorited, _ = GetTotalFavoritedByUserID(user.UserID)
	if result.RowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	//更新 redis
	if err = AddUserInfoByUserIDFromCacheToRedis(user); err != nil {
		return nil, err
	}
	return

}

// GetUserListByUserIDList 根据 UserIDList 获取对应的用户列表
func GetUserListByUserIDList(UserIDList []uint64) ([]model.User, error) {
	//查询redis
	userList, notInCache, err := GetUserListByUserIDListFromRedis(UserIDList)
	if err != nil && err.Error() != "Not found in cache" {
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
		uniqueUserList[idx].FollowCount, _ = GetFollowCountByUserID(user.UserID)
		uniqueUserList[idx].FollowerCount, _ = GetFollowerCountByUserID(user.UserID)
		uniqueUserList[idx].FavoriteCount, _ = GetFavoriteCountByUserID(user.UserID)
		uniqueUserList[idx].TotalFavorited, _ = GetTotalFavoritedByUserID(user.UserID)
		mapUserIDToUser[user.UserID] = uniqueUserList[idx]
	}
	//更新 redis
	if err = AddUserListByUserIDListsToRedis(uniqueUserList); err != nil {
		return nil, err
	}
	//后续操作
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

//// GetUserListByUserIDs 根据UserIDs获取对应的用户列表
//func GetUserListByUserIDs(UserIDs []uint64, userList *[]model.User) (err error) {
//	var uniqueUserList []model.User
//	result := global.DB.Where("user_id in ?", UserIDs).Find(&uniqueUserList)
//	if result.Error != nil {
//		err = errors.New("query GetUserListByUserIDs error")
//		return
//	}
//	// 针对查询结果建立映射关系
//	mapUserIDToUser := make(map[uint64]model.User)
//	*userList = make([]model.User, len(UserIDs))
//	for idx, user := range uniqueUserList {
//		mapUserIDToUser[user.UserID] = uniqueUserList[idx]
//	}
//	// 构造返回值
//	for idx, userID := range UserIDs {
//		(*userList)[idx] = mapUserIDToUser[userID]
//	}
//	return
//}

//// IsUserIDExist 判断 userID 是否有效
//func IsUserIDExist(userID uint64) bool {
//	var count int64
//	global.DB.Model(&model.User{}).Where("user_id = ?", userID).Count(&count)
//	return count != 0
//}
