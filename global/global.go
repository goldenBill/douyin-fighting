package global

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	DB                   *gorm.DB
	REDIS                *redis.Client
	CONTEXT              = context.Background()
	FILE_TYPE_MAP        sync.Map
	ID_GENERATOR         *sonyflake.Sonyflake
	AUTO_CREATE_DB             = false                   //是否自动生成数据库
	MAX_USERNAME_LENGTH        = 32                      // 用户名最大长度
	MIN_PASSWORD_PATTERN       = "^[_a-zA-Z0-9]{6,32}$"  //密码格式
	JWT_SigningKey             = "douyin-fighting-redis" //初始化全局签名
	START_TIME                 = "2022-05-21 00:00:01"   //固定启动时间，保证生成 ID 唯一性
	FEED_NUM                   = 2                       //每次返回视频数量
	VIDEO_ADDR                 = "./public/video/"       //视频存放位置
	COVER_ADDR                 = "./public/cover/"       //封面存放位置
	MAX_FILE_SIZE        int64 = 10 << 20                // 上传文件大小限制为10MB
	MAX_TITLE_LENGTH           = 140                     // 视频描述最大长度
	MAX_COMMENT_LENGTH         = 300                     // 评论最大长度
	WHITELIST_VIDEO            = map[string]bool{".mp4": true, ".avi": true, ".wmv": true, ".mpeg": true,
		".mov": true, ".flv": true, ".rmvb": true, ".3gb": true, ".vob": true, ".m4v": true}
)

var (
	MAX_RETRIES      int           = 1000
	FAVORITE_EXPIRE  time.Duration = 10 * time.Minute
	FOLLOW_EXPIRE    time.Duration = 10 * time.Minute
	USER_INFO_EXPIRE time.Duration = 10 * time.Minute
	VIDEO_EXPIRE     time.Duration = 10 * time.Minute
)
