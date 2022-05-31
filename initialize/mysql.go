package initialize

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func MySQL() {

	username := "root"     //账号
	password := "huangshm" //密码
	host := "localhost"    //数据库地址，可以是Ip或者域名
	port := 3306           //数据库端口
	dbName := "douyin"     //数据库名
	//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)

	// 配置Gorm连接到MySQL
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN
		DefaultStringSize:         256,   // string 类型字段的默认长度
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Millisecond * 20, // 慢 SQL 阈值
			LogLevel:                  logger.Warn,           // 日志级别
			IgnoreRecordNotFoundError: true,                  // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,                 // 禁用彩色打印
		},
	)
	if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
		Logger: newLogger,
	}); err == nil {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(100) // 设置数据库最大连接数
		sqlDB.SetMaxIdleConns(10)  // 设置上数据库最大闲置连接数
		global.DB = db
	} else {
		panic("connect server failed")
	}

	if global.AUTO_CREATE_DB {
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.User{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Video{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Favorite{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Comment{})
		global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Follow{})
	}

}
