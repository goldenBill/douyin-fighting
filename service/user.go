package service

import (
	"errors"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
)

// Register : 用户注册
func Register(username string, password string) (user *dao.User, err error) {
	//判断用户名是否存在
	rowsAffected := global.GVAR_DB.Debug().Where("name = ?", username).Limit(1).Find(&user).RowsAffected
	if rowsAffected != 0 {
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
	err = global.GVAR_DB.Debug().Create(user).Error
	return
}

// Login : 用户登录
func Login(username string, password string) (user *dao.User, err error) {
	//检查用户名是否存在
	rowsAffected := global.GVAR_DB.Debug().Where("name = ?", username).Limit(1).Find(&user).RowsAffected
	if rowsAffected == 0 {
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

// UserInfoByUserID : 通过 UserID 获取用户信息
func UserInfoByUserID(userID uint64) (user *dao.User, err error) {
	//检查 userID 是否存在；若存在，获取用户信息
	rowsAffected := global.GVAR_DB.Debug().Where("user_id = ?", userID).Limit(1).Find(&user).RowsAffected
	if rowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	err = nil
	return
}

// 判断userID是否有效
func IsUserIDExist(userID uint64) bool {
	var count int64
	global.GVAR_DB.Debug().Model(&dao.User{}).Where("user_id = ?", userID).Count(&count)
	return count != 0
}
