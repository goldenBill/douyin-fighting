package initialize

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldenBill/douyin-fighting/controller"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/middleware"
)

func Router() {
	r := gin.Default()
	// public directory is used to serve static resources
	r.Static("/static", "./public")

	apiRouter := r.Group("/douyin")

	// basic apis
	apiRouter.GET("/feed/", controller.Feed)
	apiRouter.POST("/user/register/", controller.Register)
	apiRouter.POST("/user/login/", controller.Login)
	apiRouter.GET("/publish/list/", controller.PublishList)

	// extra apis - I
	apiRouter.GET("/favorite/list/", controller.FavoriteList)
	apiRouter.GET("/comment/list/", controller.CommentList)

	// extra apis - II
	apiRouter.GET("/relation/follow/list/", controller.FollowList)
	apiRouter.GET("/relation/follower/list/", controller.FollowerList)

	// 需要 token 验证的路由
	authed := apiRouter.Group("/")
	authed.Use(middleware.JWT())
	{
		// basic apis
		authed.GET("/user/", controller.UserInfo)

		// extra apis - I
		authed.POST("/favorite/action/", controller.FavoriteAction)
		authed.POST("/comment/action/", controller.CommentAction)

		// extra apis - II
		authed.POST("/relation/action/", controller.RelationAction)
	}

	authed2 := apiRouter.Group("/")
	authed2.Use(middleware.JWT())
	authed2.Use(middleware.FileCheck())
	{
		// basic apis
		authed2.POST("/publish/action/", controller.Publish)
	}

	r.Run(fmt.Sprintf("%s:%d", global.CONFIG.GinConfig.Host, global.CONFIG.GinConfig.Port))
}
