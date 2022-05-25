package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
)

// Register 用户注册
func Register(username string, password string) (user *dao.User, err error) {
	//判断用户名是否存在
	result := global.GVAR_DB.Where("name = ?", username).Limit(1).Find(&user)
	if result.RowsAffected != 0 {
		err = errors.New("user already exists")
		return
	}
	//接收姓名
	user.Name = username
	//对明文密码加密
	user.Password = util.BcryptHash(password)
	//生成增长的 userID
	user.UserID, _ = global.GVAR_ID_GENERATOR.NextID()
	//存储到数据库
	err = global.GVAR_DB.Create(user).Error
	return
}

// Login 用户登录
func Login(username string, password string) (user *dao.User, err error) {
	//检查用户名是否存在
	result := global.GVAR_DB.Where("name = ?", username).Limit(1).Find(&user)
	if result.RowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	//检查密码是否正确
	if ok := util.BcryptCheck(password, user.Password); !ok {
		err = errors.New("wrong password")
		return
	}
	err = nil
	return
}

// UserInfoByUserID 通过 UserID 获取用户信息
func UserInfoByUserID(userID uint64) (user *dao.User, err error) {
	//检查 userID 是否存在；若存在，获取用户信息
	result := global.GVAR_DB.Where("user_id = ?", userID).Limit(1).Find(&user)
	if result.RowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	err = nil
	return
}

// IsUserIDExist 判断 userID 是否有效
func IsUserIDExist(userID uint64) bool {
	var count int64
	global.GVAR_DB.Model(&dao.User{}).Where("user_id = ?", userID).Count(&count)
	return count != 0
}

// GetUserListByUserIDs 根据UserIDs获取对应的用户列表
func GetUserListByUserIDs(UserIDs []uint64) (userList []dao.User, err error) {
	var uniqueUserList []dao.User
	result := global.GVAR_DB.Where("user_id in ?", UserIDs).Find(&uniqueUserList)
	if result.Error != nil {
		err = errors.New("query GetUserListByUserIDs error")
	}
	mapUserIDToUser := make(map[uint64]*dao.User)
	for idx, user := range uniqueUserList {
		mapUserIDToUser[user.UserID] = &uniqueUserList[idx]
	}
	userList = make([]dao.User, len(UserIDs))
	for idx, userID := range UserIDs {
		userList[idx] = *mapUserIDToUser[userID]
	}
	return
}
