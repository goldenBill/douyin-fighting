package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller"
	"github.com/goldenBill/douyin-fighting/util"
	"net/http"
)

// 定义中间
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
		//if !service.IsUserIDExist(userID) {
		//	c.JSON(http.StatusForbidden, controller.Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		//	c.Abort()
		//	return
		//}
		// 保存userID到Context的key中，可以通过Get()取
		c.Set("UserID", userID)

		// 执行函数
		c.Next()
	}
}
