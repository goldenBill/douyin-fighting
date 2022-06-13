package initialize

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/goldenBill/douyin-fighting/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func MySQL() {

	username := global.CONFIG.MySQLConfig.Username // 账号
	password := global.CONFIG.MySQLConfig.Password // 密码
	host := global.CONFIG.MySQLConfig.Host         // 数据库地址，可以是Ip或者域名
	port := global.CONFIG.MySQLConfig.Port         // 数据库端口
	dbName := global.CONFIG.MySQLConfig.DBname     // 数据库名
	// dsn := "用户名:密码@tcp(地址:端口)/数据库名"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)

	// 配置Gorm连接到MySQL
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN
		DefaultStringSize:         256,   // string 类型字段的默认长度
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Millisecond * 0, // 慢 SQL 阈值
			LogLevel:                  logger.Info,          // 日志级别
			IgnoreRecordNotFoundError: true,                 // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,                // 禁用彩色打印
		},
	)
	if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
		Logger: newLogger,
	}); err == nil {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(global.CONFIG.MySQLConfig.MaxOpenConns) // 设置数据库最大连接数
		sqlDB.SetMaxIdleConns(global.CONFIG.MySQLConfig.MaxIdleConns) // 设置上数据库最大闲置连接数
		global.DB = db
	} else {
		panic("connect server failed")
	}

	// 自动生成对应的数据库表
	if global.AUTO_CREATE_DB {
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model.User{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model.Video{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model.Favorite{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model.Comment{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&model.Follow{})
	}

}
