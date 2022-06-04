# 极简版抖音
## 1. 配置
- 下载[ffpmeg](https://ffmpeg.org/)，并可以在命令行运行
- 在config/config.yml中配置信息
```yaml
gin:
  host: 0.0.0.0
  port: 8080

mysql:
  host: localhost
  port: 3306
  username: root
  password: huangshm
  db_name: douyin
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379
  password:
  db: 1
  pool_size: 100
```
- 自动生成数据库（默认开启，不会覆盖数据）
```go
package global

AUTO_CREATE_DB bool = true     //是否自动生成数据库
```

## 2. 用户注册登录接口

### 主要文件及功能说明
```
douyin-fighting/
│
├─controller
│     user.go           用户注册登录
│
├─model
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

### 优化点
更新数据，先写数据库，再删缓存

## 3. 视频上传发布接口

### 主要文件及功能说明
```
douyin-fighting/
│
├─controller
│     feed.go           视频推送
│     publish.go        视频发布以及信息
│
├─model
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
├─model
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

## 5. 关注接口

### 主要文件及功能说明
```
douyin-fighting/
│
├─controller
│     relation.go
│
├─model
│     follow.go
│
└─service
      follow.go
```
### 文档说明