package global

import (
	"github.com/jmoiron/sqlx"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
)

var (
	GVAR_SQLX_DB        *sqlx.DB
	GVAR_DB             *gorm.DB
	GVAR_ID_GENERATOR   *sonyflake.Sonyflake
	GVAR_JWT_SigningKey string = "douyin-fighting"     //初始化全局签名
	GVAR_START_TIME     string = "2022-05-21 00:00:01" //固定启动时间，保证生成 Id 唯一性
	GVAR_FEED_NUM       int    = 30
)
