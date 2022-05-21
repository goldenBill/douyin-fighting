package service

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
	"gorm.io/gorm"
	"time"
)

// User : 定义 user 服务的结构体
type User struct{}

// UserClaims : 自定义 claims 类型
type UserClaims struct {
	ID     uint64
	UserID uint64
	Name   string
	jwt.RegisteredClaims
}

// Register : 用户注册
func (User *User) Register(user *dao.User) error {
	var userTmp dao.User
	//判断用户名是否存在
	err := global.GVAR_DB.Debug().Where("name = ?", user.Name).First(&userTmp).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user already exists")
	}
	//对明文密码加密
	user.Password = util.BcryptHash(user.Password)
	//生成增长的 userId
	user.UserId, _ = global.GVAR_ID_GENERATOR.NextID()
	//存储到数据库
	err = global.GVAR_DB.Debug().Create(user).Error
	return err
}

// Login : 用户登录
func (User *User) Login(user *dao.User) (*dao.User, error) {
	var userTmp dao.User
	//检查用户名是否存在
	err := global.GVAR_DB.Debug().Where("name = ?", user.Name).First(&userTmp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &userTmp, errors.New("username does not exist")
	}
	//检查密码是否正确
	if ok := util.BcryptCheck(user.Password, userTmp.Password); !ok {
		return &userTmp, errors.New("wrong password")
	}
	return &userTmp, nil
}

// UserInfo : 获取用户信息
func (User *User) UserInfo(userId uint64) (*dao.User, error) {
	var userTmp dao.User
	//检查 userId 是否存在；若存在，获取用户信息
	err := global.GVAR_DB.Where("user_id = ?", userId).First(&userTmp).Error
	return &userTmp, err
}

// GenerateToken : 生成 token
func (User *User) GenerateToken(user *dao.User) (string, error) {
	//获取全局签名
	mySigningKey := []byte(global.GVAR_JWT_SigningKey)
	//配置 userClaims ,并生成 token
	claims := UserClaims{
		user.Id,
		user.UserId,
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
func (User *User) ParseToken(tokenString string) (*jwt.Token, error) {
	//获取全局签名
	mySigningKey := []byte(global.GVAR_JWT_SigningKey)
	//解析 token 信息
	return jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
}
