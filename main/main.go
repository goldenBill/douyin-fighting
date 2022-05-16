package main

import "github.com/goldenBill/douyin-fighting/config"

func main() {
	config.InitMySQL()
	config.InitRouter()
}
