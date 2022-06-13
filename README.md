<h1 align="center" style="font-size:50px">douyin-fighting</h1>
<div align=center>
<img src="https://img.shields.io/badge/golang-1.18-blue"/>
<img src="https://img.shields.io/badge/gin-1.7.7-yellowgreen"/>
<img src="https://img.shields.io/badge/go--redis-8.11.5-brightgreen"/>
<img src="https://img.shields.io/badge/gorm-1.23.5-red"/>
<img src="https://img.shields.io/badge/viper-1.12.0-orange"/>
</div>

# 1. 基本介绍
## 1.1 项目架构
本项目采用 MVC 分层设计模型分离模型层、视图层和控制层，从而降低代码的耦合度，提高项目的可维护性。

使用 Gin 作为 Web 框架，Redis 作为缓存框架，MySQL 作为持久层框架。
```
         ┌─────────┐      ┌─────────┐      ┌─────────┐
  ──req──►         ├──────►         ├──────►         │
         │   Gin   │      │  Redis  │      │  MySQL  │
  ◄─resp─┤         ◄──────┤         ◄──────┤         │
         └─────────┘      └─────────┘      └─────────┘
```
## 1.2 功能介绍
- 视频：视频推送、视频投稿、发布列表
- 用户：用户注册、用户登录、用户信息
- 点赞：点赞操作、点赞列表
- 评论：评论操作、评论列表
- 关注：关注操作、关注列表、粉丝列表

## 1.3 技术选型
- Gin
- Gorm
- MySQL
- Redis
- Viper
- ffmpeg

## 1.4 使用说明
- 安装配置Golang、MySQL、Redis 和 ffmpeg
- 启动服务
```bash
# 克隆 github 项目最新版本
git clone -b redis-enhance https://github.com/goldenBill/douyin-fighting.git --depth=1

# 编译运行项目
cd douyin-fighting
chmod u+x run.sh
./run.sh
```

## 1.5 演示地址
```shell
http://47.113.191.56:8080/
```

# 2. 目录结构
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

# 3. 文档说明

详细说明请参考
[2022字节跳动后端青训营抖音项目文档](https://gjj3nncz08.feishu.cn/docx/doxcnJGyQHB30zcOkkT3CkTC6Oc)