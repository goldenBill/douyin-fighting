package global

import (
	"github.com/jmoiron/sqlx"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
	"sync"
)

var (
	GVAR_DB              *gorm.DB
	GVAR_SQLX_DB         *sqlx.DB
	GVAR_ID_GENERATOR    *sonyflake.Sonyflake
	GVAR_AUTO_CREATE_DB  bool            = false                 //是否自动生成数据库
	GVAR_JWT_SigningKey  string          = "douyin-fighting"     //初始化全局签名
	GVAR_START_TIME      string          = "2022-05-21 00:00:01" //固定启动时间，保证生成 ID 唯一性
	GVAR_FEED_NUM        int             = 1
	GVAR_VIDEO_ADDR      string          = "./public/video/"
	GVAR_COVER_ADDR      string          = "./public/cover/"
	GVAR_WHITELIST_VIDEO map[string]bool = map[string]bool{".mp4": true, ".avi": true, ".wmv": true, ".mpeg": true, ".mov": true, ".flv": true, ".rmvb": true, ".3gb": true, ".vob": true, ".m4v": true}
	GVAR_FILE_TYPE_MAP   sync.Map
	GVAR_FILE_MAX_SIZE   int64 = 200 * 1024 * 1024 // 200MB
)
