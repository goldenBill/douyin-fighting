package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller"
	"github.com/goldenBill/douyin-fighting/middleware"
)

func InitRouter() {
	r := gin.Default()
	// public directory is used to serve static resources
	r.Static("/static", "./public")

	apiRouter := r.Group("/douyin")

	// basic apis
	apiRouter.GET("/feed/", controller.Feed)
	apiRouter.POST("/user/register/", controller.Register)
	apiRouter.POST("/user/login/", controller.Login)

	// extra apis - I
	apiRouter.GET("/comment/list/", controller.CommentList)

	// 需要 token 验证的路由
	authed := apiRouter.Group("/")
	authed.Use(middleware.JWT())
	{
		// basic apis
		authed.GET("/user/", controller.UserInfo)
		authed.POST("/publish/action/", controller.Publish)
		authed.GET("/publish/list/", controller.PublishList)

		// extra apis - I
		authed.POST("/favorite/action/", controller.FavoriteAction)
		authed.GET("/favorite/list/", controller.FavoriteList)
		authed.POST("/comment/action/", controller.CommentAction)

		// extra apis - II
		authed.POST("/relation/action/", controller.RelationAction)
		authed.GET("/relation/follow/list/", controller.FollowList)
		authed.GET("/relation/follower/list/", controller.FollowerList)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
