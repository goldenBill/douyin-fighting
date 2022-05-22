package controller

import "github.com/goldenBill/douyin-fighting/service"

var (
	userService  = service.User{}  //开启 userService 服务
	videoService = service.Video{} //开启 videoService 服务
)
