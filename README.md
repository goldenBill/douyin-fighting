# 极简版抖音
## 1. 配置
- 下载[ffpmeg](https://ffmpeg.org/)，并可以在命令行运行
- 修改config/mysql.go中数据库**用户名**和**密码**
```go
//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
dsn := "用户名:密码@tcp(127.0.0.1:3306)/douyin?charset=utf8&parseTime=true"
```
