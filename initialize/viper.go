package initialize

import (
	"github.com/fsnotify/fsnotify"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/spf13/viper"
	"log"
)

func Viper() {
	// 设置配置文件类型和路径
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./config/config.yml")
	// 读取配置信息
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("获取配置文件错误")
	}
	// 将读取到的配置信息反序列化到全局 CONFIG 中
	err = viper.Unmarshal(&global.CONFIG)
	if err != nil {
		log.Panic("viper反序列化错误")
	}
	log.Println(global.CONFIG)
	// 监视配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置文件被修改：", e.Name)
	})
}
