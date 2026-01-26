# UC-Server (多租户用户中心)

基于 Gin + MySQL + Redis 开发的多租户用户中心服务，提供用户认证、授权、多租户管理等核心功能。

## 项目架构

本项目采用经典的分层架构设计，遵循 Go 语言最佳实践和 Clean Architecture 原则。

## 目录结构

```text
uc-server/
├── cmd/                           # 应用程序入口
│   └── server/                    # 服务端主程序
│       └── main.go               # 程序启动入口，初始化配置、数据库、路由等
│
├── config/                        # 配置模块
│   ├── config.go                 # 配置结构体定义和加载逻辑
│   └── config.yaml               # 应用配置文件（数据库、Redis、JWT等配置）
│
├── internal/                      # 内部私有代码（不对外暴露）
│   ├── app/                      # 应用程序核心
│   │   ├── app.go               # 应用初始化、路由注册、中间件配置
│   │   └── context.go           # 自定义上下文，扩展 Gin Context
│   │
│   ├── dto/                      # 数据传输对象（Data Transfer Object）
│   │   ├── auth.go          # 认证相关 DTO（登录、注册请求/响应）
│   │   └── user.go          # 用户相关 DTO（用户信息传输对象）
│   │
│   ├── handler/                  # HTTP 处理器层（Controller）
│   │   ├── auth.go              # 认证接口（登录、注册、登出等）
│   │   └── user.go              # 用户管理接口（查询、更新、删除等）
│   │
│   ├── middleware/               # 中间件
│   │   ├── body_size_limit.go   # 请求体大小限制中间件
│   │   ├── cors.go              # 跨域资源共享（CORS）中间件
│   │   ├── jwt.go               # JWT 认证中间件
│   │   ├── logging.go           # 请求日志记录中间件
│   │   ├── rate_limit.go        # 请求限流中间件
│   │   └── recovery.go          # Panic 恢复中间件
│   │
│   ├── model/                    # 数据模型层（Domain Model）
│   │   ├── base.go              # 基础模型（公共字段：ID、创建时间、更新时间等）
│   │   ├── constants.go         # 常量定义（状态码、用户角色等）
│   │   ├── employee.go          # 员工模型
│   │   ├── tenant.go            # 租户模型
│   │   ├── tenant_user.go       # 租户-用户关联模型（多对多关系）
│   │   └── user.go              # 用户模型
│   │
│   ├── pkg/                      # 内部公共包
│   │   ├── cache/               # 缓存层
│   │   │   ├── auth.go    # 认证相关缓存（Token、会话等）
│   │   │   ├── cache.go         # 缓存接口定义
│   │   │   ├── redis.go         # Redis 客户端初始化和操作
│   │   │   └── user.go    # 用户缓存
│   │   │
│   │   ├── database/            # 数据库层
│   │   │   └── mysql.go         # MySQL 连接初始化（GORM）
│   │   │
│   │   ├── jwt/                 # JWT 工具
│   │   │   └── jwt.go           # JWT 生成、解析、验证
│   │   │
│   │   ├── response/            # 响应处理
│   │   │   └── response.go      # 统一响应格式封装
│   │   │
│   │   ├── snowflake/           # 分布式ID生成器
│   │   │   ├── snowflake.go     # Snowflake 算法实现
│   │   │   └── snowflake_test.go # Snowflake 单元测试
│   │   │
│   │   ├── utils/               # 工具函数
│   │   │   └── token.go         # Token 相关工具函数
│   │   │
│   │   └── validation/          # 数据验证
│   │       ├── password.go      # 密码强度验证
│   │       ├── password_test.go # 密码验证测试
│   │       ├── validator_error.go      # 验证器错误处理
│   │       ├── validator_error_test.go # 验证器错误测试
│   │       └── README.md        # 验证模块说明文档
│   │
│   ├── repository/               # 数据访问层（Repository Pattern）
│   │   ├── base.go   # 基础仓储（通用 CRUD 操作）
│   │   ├── tenant_user.go # 租户用户关联仓储
│   │   └── user.go   # 用户仓储（用户数据访问）
│   │
│   └── service/                  # 业务逻辑层（Service Layer）
│       ├── auth.go      # 认证服务（登录、注册、Token 管理）
│       └── user.go      # 用户服务（用户业务逻辑）
│
├── scripts/                      # 脚本文件
│   └── init.sql                 # 数据库初始化脚本
│
├── deploy/                       # 部署相关
│   └── Dockerfile               # Docker 镜像构建文件
│
├── .air.toml                     # Air 热重载配置
├── .gitignore                    # Git 忽略文件配置
├── go.mod                        # Go 模块依赖管理
├── go.sum                        # Go 模块依赖校验
├── Makefile                      # 自动化构建脚本
└── README.md                     # 项目说明文档
```

## 技术栈

- **Web 框架**: Gin
- **数据库**: MySQL (GORM)
- **缓存**: Redis
- **认证**: JWT
- **ID 生成**: Snowflake 算法
- **热重载**: Air

## 架构分层说明

### 1. Handler 层（HTTP 处理器）

负责接收 HTTP 请求，参数验证，调用 Service 层处理业务逻辑，返回响应。

### 2. Service 层（业务逻辑）

封装核心业务逻辑，处理复杂的业务流程，协调多个 Repository 完成业务操作。

### 3. Repository 层（数据访问）

负责数据持久化操作，封装数据库查询逻辑，为 Service 层提供数据访问接口。

### 4. Model 层（数据模型）

定义数据库表结构对应的实体模型，使用 GORM 标签映射。

### 5. DTO 层（数据传输）

定义 API 请求和响应的数据结构，与 Model 分离，便于版本控制和数据转换。

### 6. Middleware 层（中间件）

提供横切关注点功能，如认证、日志、限流、跨域等。

## 多租户设计

本项目采用**共享数据库、共享 Schema** 的多租户模式：

- 通过 `tenant_id` 字段区分不同租户的数据
- `tenant_user` 表实现用户与租户的多对多关联
- 支持一个用户属于多个租户
- Repository 层自动注入租户上下文，确保数据隔离

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 安装依赖

```bash
go mod download
```

### 配置文件

复制并修改配置文件：

```bash
cp config/config.yaml.example config/config.yaml
```

### 初始化数据库

```bash
mysql -u root -p < scripts/init.sql
```

### 运行项目

```bash
# 开发模式（带热重载）
air

# 或使用 Makefile
make run

# 生产模式
make build
./bin/server
```

## 开发规范

- 遵循 Go 语言官方代码规范
- 使用 GORM 进行数据库操作
- 统一使用 `internal/pkg/response` 封装 API 响应
- 敏感操作必须经过 JWT 认证中间件
- 所有数据访问必须通过 Repository 层
- 业务逻辑封装在 Service 层

## 许可证

MIT License
