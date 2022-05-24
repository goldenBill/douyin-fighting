package initialize

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitMySQL() {

	username := "root"          //账号
	password := "Zxy13525812.." //密码
	host := "127.0.0.1"         //数据库地址，可以是Ip或者域名
	port := 3306                //数据库端口
	dbName := "douyin"          //数据库名
	//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)

	// 配置sqlx连接到MySQL
	if db, err := sqlx.Open("mysql", dsn); err == nil {
		db.SetConnMaxLifetime(100) // 设置数据库最大连接数
		db.SetMaxIdleConns(10)     // 设置上数据库最大闲置连接数
		global.GVAR_SQLX_DB = db
	} else {
		panic("connect server failed")
	}

	// 配置Gorm连接到MySQL
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN
		DefaultStringSize:         256,   // string 类型字段的默认长度
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
	}

}
