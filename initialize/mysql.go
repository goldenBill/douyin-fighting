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

func InitMySQL() {

	username := "root"     //账号
	password := "huangshm" //密码
	host := "127.0.0.1"    //数据库地址，可以是Ip或者域名
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
		global.GVAR_DB = db
	} else {
		panic("connect server failed")
	}

	if global.GVAR_AUTO_CREATE_DB {
		global.GVAR_DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.User{})
		global.GVAR_DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Video{})
		global.GVAR_DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Favorite{})
		global.GVAR_DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Comment{})
		global.GVAR_DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&dao.Follow{})
	}

}
