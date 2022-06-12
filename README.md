<h1 align="center" style="font-size:50px">douyin-fighting</h1>
<div align=center>
<img src="https://img.shields.io/badge/golang-1.18-blue"/>
<img src="https://img.shields.io/badge/gin-1.7.7-yellowgreen"/>
<img src="https://img.shields.io/badge/go--redis-8.11.5-brightgreen"/>
<img src="https://img.shields.io/badge/gorm-1.23.5-red"/>
<img src="https://img.shields.io/badge/viper-1.12.0-orange"/>
</div>

# 1.项目架构
本项目采用 MVC 分层设计模型分离模型层、视图层和控制层，从而降低代码的耦合度，提高项目的可维护性。

使用 Gin 作为 Web 框架，Redis 作为缓存框架，MySQL 作为持久层框架。
```
         ┌─────────┐      ┌─────────┐      ┌─────────┐
  ──req──►         ├──────►         ├──────►         │
         │   Gin   │      │  Redis  │      │  MySQL  │
  ◄─resp─┤         ◄──────┤         ◄──────┤         │
         └─────────┘      └─────────┘      └─────────┘
```
# 2. 使用说明
安装配置 MySQL、Redis、ffmpeg 以及 go 运行环境，确保可以在命令行运行
```shell
./run.sh
```
# 3. 项目技术栈
- Gin
- Gorm
- MySQL
- Redis
- Viper
- ffmpeg

# 4. 目录结构
```
├─ config（配置文件信息以及Viper管理文件）
│    ├─ config.go
│    └─ config.yml
├─ controller（处理客户端请求的控制层）
│    ├─ comment.go
│    ├─ common.go
│    ├─ favorite.go
│    ├─ feed.go
│    ├─ publish.go
│    ├─ relation.go
│    └─ user.go
├─ global（全局变量）
│    └─ global.go
├─ initialize（初始化操作）
│    ├─ global.go
│    ├─ mysql.go
│    ├─ redis.go
│    ├─ router.go
│    └─ viper.go
├─ main（主程序入口）
│    └─ main.go
├─ middleware（中间件）
│    ├─ filecheck.go（上传文件合法性检查）
│    └─ jwt.go（权限校验）
├─ model（模型层）
│    ├─ comment.go
│    ├─ favorite.go
│    ├─ follow.go
│    ├─ user.go
│    └─ video.go
├─ service（业务逻辑层，包含数据库操作）
│    ├─ comment.go
│    ├─ comment_redis.go
│    ├─ common.go
│    ├─ favorite.go
│    ├─ favorite_redis.go
│    ├─ follow.go
│    ├─ follow_redis.go
│    ├─ new_test.go
│    ├─ user.go
│    ├─ user_redis.go
│    ├─ video.go
│    └─ video_redis.go
└─ util（工具类）
       ├─ encryption.go（加密）
       ├─ file.go（文件路径生成）
       ├─ jwt.go（权限校验）
       └─ video.go（视频处理）
```
# 5. 功能介绍
- 视频发布推送
- 用户注册登录
- 点赞评论功能
- 关注粉丝功能

# 6. 技术说明
1. 权限生成和验证
  - 基于 token 的无状态性质，我们采用 token 来实现权限的记录以及会话的保持。
  - 将 token 验证的模块安排在中间件中，对相应功能进行权限检查
2. 修改和读取数据
  - 在修改数据操作中，针对 count 计数数据的 “每次仅加减一” 的特性，采用先修改数据库再更新缓存的策略
  - 在读取数据操作中，我们采用先读缓存数据，若不存在则从数据库读取数据，并写入缓存的策略
3. 持久层数据
  - 采用 snowflake 算法生成递增的主键
  - 由于每次查询持久层数据的通信开销过大，本项目采用批量操作来减少持久层访问
  - 为了保护用户密码的安全性，采用不可逆加密算法对用户密码进行加密后再存储到持久化层
4. 缓存数据
  - 将请求中常用的 follow_count, follower_count, total_favorited, favorite_count 记录到 Redis 缓存层中，从而减少对数据库的查询，提高系统的响应速度
  - 我们利用 Lua 脚本和 pipeline 多命令执行来保证 Redis 操作的原子性
  - 设置随机的过期时间，减少缓存雪崩现象的发生
5. go 特性使用
  - 针对某些并行操作，我们使用 go 协程来加速处理
6. 日志记录
  - 开启 Gin 框架和 Gorm 框架的日志记录功能
7. 配置文件
  - 使用 yaml 文件记录相关的服务配置信息，并使用 Viper 对配置文件进行读取解析和动态监控