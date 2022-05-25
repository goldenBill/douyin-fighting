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
	// 创建白名单类型
	global.GVAR_FILE_TYPE_MAP.Store("0000002066747970", ".mp4")
	global.GVAR_FILE_TYPE_MAP.Store("0000001c66747970", ".mp4")
	global.GVAR_FILE_TYPE_MAP.Store("0000001866747970", ".mp4")
	global.GVAR_FILE_TYPE_MAP.Store("52494646", ".avi")
	global.GVAR_FILE_TYPE_MAP.Store("3026b2758e66cf11a6d9", ".wmv")
	global.GVAR_FILE_TYPE_MAP.Store("000001BA47000001B3", ".mpeg")
	global.GVAR_FILE_TYPE_MAP.Store("6D6F6F76", ".mov")
	global.GVAR_FILE_TYPE_MAP.Store("464c5601050000000900", ".flv")
	global.GVAR_FILE_TYPE_MAP.Store("2e524d46000000120001", ".rmvb")
	global.GVAR_FILE_TYPE_MAP.Store("667479703367", ".3gb")
	global.GVAR_FILE_TYPE_MAP.Store("000001BA", ".vob")
	global.GVAR_FILE_TYPE_MAP.Store("00000020667479704D34412000000000", ".m4v")

}
