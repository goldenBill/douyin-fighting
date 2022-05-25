package initialize

import (
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
	"github.com/sony/sonyflake"
	"time"
)

func InitGlobal() {
	//初始化 ID 生成器
	startTime, _ := time.Parse("2006-01-02 15:04:05", global.GVAR_START_TIME)
	global.GVAR_ID_GENERATOR = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: startTime,
	})
	//创建video存放目录
	util.CheckPathAndCreate(global.GVAR_VIDEO_ADDR)
	util.CheckPathAndCreate(global.GVAR_COVER_ADDR)

}
