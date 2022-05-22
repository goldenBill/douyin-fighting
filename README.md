# 极简版抖音
## 1. 配置
- 下载[ffpmeg](https://ffmpeg.org/)，并可以在命令行运行
- 修改config/mysql.go中数据库**用户名**和**密码**
```go
//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
dsn := "root:huangshm@tcp(127.0.0.1:3306)/douyin?charset=utf8mb4&parseTime=true&loc=Local"
```
## 2. user_info 负责部分
### 主要涉及部分
```
douyin-fighting/
│
├─controller
│     common.go
│     service.go
│     user.go
│
├─dao
│     user.go
│
└─service
      user.go
```
### 文档说明
- 对密码进行加密存入数据库加密
- 唯一的userID
- 生成token权鉴
- 优化项目结构