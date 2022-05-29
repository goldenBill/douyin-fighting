package main

import "github.com/goldenBill/douyin-fighting/initialize"

func main() {
	initialize.InitRedis()
	initialize.InitGlobal() //初始化全局变量
	initialize.InitMySQL()  //初始化 MySQL 连接
	initialize.InitRouter() //初始化 GinRouter
}
