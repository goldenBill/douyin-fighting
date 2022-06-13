package util

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"time"
)

type UserClaims struct {
	UserID uint64
	Name   string
	jwt.RegisteredClaims
}

// GenerateToken 生成 token
func GenerateToken(user *model.User) (string, error) {
	// 获取全局签名
	mySigningKey := []byte(global.CONFIG.JWTConfig.SigningKey)
	// 配置 userClaims ,并生成 token
	claims := UserClaims{
		user.UserID,
		user.Name,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

// ParseToken 解析 token
func ParseToken(tokenString string) (*UserClaims, error) {
	// 获取全局签名
	mySigningKey := []byte(global.CONFIG.JWTConfig.SigningKey)
	// 解析 token 信息
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		return nil, err
	} else if token == nil {
		return nil, errors.New("token is invalid")
	}
	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, errors.New("token is invalid")
}
