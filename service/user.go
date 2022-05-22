package service

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
	"time"
)

type User struct {
}

type UserClaims struct {
	ID     uint64
	UserID uint64
	Name   string
	jwt.RegisteredClaims
}

// Register : 用户注册
func (*User) Register(username string, password string) (user *dao.User, err error) {
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
func (*User) Login(username string, password string) (user *dao.User, err error) {
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

// GenerateToken : 生成 token
func (*User) GenerateToken(user *dao.User) (string, error) {
	//获取全局签名
	mySigningKey := []byte(global.GVAR_JWT_SigningKey)
	//配置 userClaims ,并生成 token
	claims := UserClaims{
		user.ID,
		user.UserID,
		user.Name,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

// ParseToken : 解析 token
func (*User) ParseToken(tokenString string) (*jwt.Token, error) {
	//获取全局签名
	mySigningKey := []byte(global.GVAR_JWT_SigningKey)
	//解析 token 信息
	return jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
}

// GetIDFromToken : 解析 token 获取 UserID
func (*User) GetIDFromToken(tokenString string) (uint64, error) {
	//获取全局签名
	mySigningKey := []byte(global.GVAR_JWT_SigningKey)
	//解析 token 信息
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	claims := token.Claims.(*UserClaims)
	return claims.ID, err
}

// UserInfoByUserID : 通过 UserID 获取用户信息
func (*User) UserInfoByUserID(userID uint64) (user *dao.User, err error) {
	//检查 userID 是否存在；若存在，获取用户信息
	rowsAffected := global.GVAR_DB.Debug().Where("user_id = ?", userID).Limit(1).Find(&user).RowsAffected
	if rowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	err = nil
	return
}

// UserInfoByID : 通过 ID 获取用户信息
func (*User) UserInfoByID(ID uint64) (user *dao.User, err error) {
	//检查 userID 是否存在；若存在，获取用户信息
	rowsAffected := global.GVAR_DB.Debug().Where("id = ?", ID).Limit(1).Find(&user).RowsAffected
	if rowsAffected == 0 {
		err = errors.New("username does not exist")
		return
	}
	err = nil
	return
}
