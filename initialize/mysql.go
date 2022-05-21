package initialize

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/goldenBill/douyin-fighting/global"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitMySQL() {
	//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
	dsn := "root:huangshm@tcp(127.0.0.1:3306)/douyin?charset=utf8mb4&parseTime=true&loc=Local"

	// 配置sqlx连接到MySQL数据库
	if db, err := sqlx.Open("mysql", dsn); err == nil {
		db.SetConnMaxLifetime(100) // 设置数据库最大连接数
		db.SetMaxIdleConns(10)     // 设置上数据库最大闲置连接数
		global.GVAR_SQLX_DB = db
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
	}
}
