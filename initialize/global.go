package initialize

import (
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/sony/sonyflake"
	"time"
)

func InitGlobal() {
	// 初始化全局签名
	global.GVAR_JWT.SigningKey = "douyin-fighting"
	// 初始化ID生成器
	global.GVAR_ID_GENERATOR = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Now(),
	})
}
