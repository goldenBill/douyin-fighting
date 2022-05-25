package global

import (
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
)

var (
	GVAR_DB              *gorm.DB
	GVAR_ID_GENERATOR    *sonyflake.Sonyflake
	GVAR_AUTO_CREATE_DB  bool   = true                   //是否自动生成数据库
	MAX_USERNAME_LENGTH  int    = 32                     // 用户名最大长度
	MIN_PASSWORD_PATTERN string = "^[_a-zA-Z0-9]{6,32}$" //密码格式
	GVAR_JWT_SigningKey  string = "douyin-fighting"      //初始化全局签名
	GVAR_START_TIME      string = "2022-05-21 00:00:01"  //固定启动时间，保证生成 ID 唯一性
	GVAR_FEED_NUM        int    = 2
	GVAR_VIDEO_ADDR      string = "./public/video/"
	GVAR_COVER_ADDR      string = "./public/cover/"
)
