# 极简版抖音
## 1. 配置
- 下载[ffpmeg](https://ffmpeg.org/)，并可以在命令行运行
- 修改config/mysql.go中数据库**用户名**和**密码**
```go
username := "root"     //账号
password := "huangshm" //密码
host := "127.0.0.1"    //数据库地址，可以是Ip或者域名
port := 3306           //数据库端口
dbName := "douyin"     //数据库名
//dsn := "用户名:密码@tcp(地址:端口)/数据库名"
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
```
## 2. 用户注册登录接口

### 主要文件及功能说明
```
douyin-fighting/
│
├─controller
│     service.go        注册数据库交互服务
│     user.go           用户注册登录
│
├─dao
│     user.go           用户数据库定义
│
└─service
      user.go           用户与数据库的交互
```
### 文档说明
- 对密码进行加密存入数据库加密
- 唯一的userID
- 生成token权鉴
- 优化项目结构

## 3. 视频上传发布接口

### 主要文件及功能说明
```
douyin-fighting/
│
├─controller
│     service.go        注册数据库交互服务
│     feed.go           视频推送
│     publish.go        视频发布以及信息
│
├─dao
│     video.go          视频数据库定义
│
└─service
      video.go          视频与数据库的交互
```
### 文档说明

## 4. 扩展接口 I

### 主要文件及功能说明

``` 
douyin-fighting/
│
├─controller
│     favorite.go		关于点赞的处理逻辑
│     comment.go		关于评论的处理逻辑
│
├─dao
│     favorite.go		点赞表数据库定义
│     comment.go		评论表数据库定义
│
└─service
      favorite.go		点赞相关功能与数据库的交互
      comment.go		评论相关功能与数据库的交互
```

### 已完成的功能

- 实现基本的点赞相关功能：
    - 点赞 / 取消点赞
    - 获取点赞列表
- 实现基本的评论功能：
    - 评论 / 删除评论
    - 获取评论列表

### 目前的问题

- 点赞、评论涉及到用户和视频，后面需要统一一下