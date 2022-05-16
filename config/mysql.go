package config

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goldenBill/douyin-fighting/dao"
	"github.com/jmoiron/sqlx"
)

func InitMySQL() {
	//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
	dsn := "用户名:密码@tcp(127.0.0.1:3306)/douyin?charset=utf8&parseTime=true"
	database, err := sqlx.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("connect server failed, err:%v\n", err)
		return
	}
	database.SetConnMaxLifetime(100) // 设置数据库最大连接数
	database.SetMaxIdleConns(10)     // 设置上数据库最大闲置连接数
	dao.DataBase = database
}
