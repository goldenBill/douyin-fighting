package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
)

// JWT 定义中间件，进行用户权限校验
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.PostForm("token")
		if tokenString == "" {
			tokenString = c.Query("token")
		}
		if tokenString == "" {
			c.JSON(http.StatusForbidden, controller.Response{StatusCode: 1, StatusMsg: "token is requested"})
			c.Abort()
			return
		}

		claims, err := util.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusForbidden, controller.Response{StatusCode: 1, StatusMsg: err.Error()})
			c.Abort()
			return
		}
		userID := claims.UserID

		// 保存userID到Context的key中，可以通过Get()取
		c.Set("UserID", userID)

		// 执行函数
		c.Next()
	}
}
