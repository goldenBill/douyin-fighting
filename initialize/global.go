package initialize

import (
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/util"
	"github.com/sony/sonyflake"
	"math/rand"
	"time"
)

func Global() {
	// 初始化随机种子
	rand.Seed(time.Now().Unix())
	// 初始化 ID 生成器
	startTime, _ := time.Parse("2006-01-02 15:04:05", global.START_TIME)
	global.ID_GENERATOR = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: startTime,
	})
	// 创建video存放目录
	util.CheckPathAndCreate(global.VIDEO_ADDR)
	util.CheckPathAndCreate(global.COVER_ADDR)
	// 创建白名单类型
	global.FILE_TYPE_MAP.Store("0000002066747970", ".mp4")
	global.FILE_TYPE_MAP.Store("0000001c66747970", ".mp4")
	global.FILE_TYPE_MAP.Store("0000001866747970", ".mp4")
	global.FILE_TYPE_MAP.Store("52494646", ".avi")
	global.FILE_TYPE_MAP.Store("3026b2758e66cf11a6d9", ".wmv")
	global.FILE_TYPE_MAP.Store("000001BA47000001B3", ".mpeg")
	global.FILE_TYPE_MAP.Store("6D6F6F76", ".mov")
	global.FILE_TYPE_MAP.Store("464c5601050000000900", ".flv")
	global.FILE_TYPE_MAP.Store("2e524d46000000120001", ".rmvb")
	global.FILE_TYPE_MAP.Store("667479703367", ".3gb")
	global.FILE_TYPE_MAP.Store("000001BA", ".vob")
	global.FILE_TYPE_MAP.Store("00000020667479704D34412000000000", ".m4v")

}
