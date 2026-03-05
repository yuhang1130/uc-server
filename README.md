# UC-Server (多租户用户中心)

基于 Gin + MySQL + Redis 开发的多租户用户中心服务，采用模块化架构和依赖注入设计，提供用户认证、授权、会话管理等核心功能。

## 项目特点

- 🏗️ **模块化架构**: 采用 DDD 分层设计，每个业务模块独立管理
- 💉 **依赖注入**: 使用 Uber Fx 框架实现自动依赖注入和生命周期管理
- 🔐 **JWT 认证**: 基于 JWT 的无状态认证，支持 Token 黑名单机制
- 🏢 **多租户支持**: 共享数据库模式，通过 tenant_id 实现数据隔离
- 📦 **缓存优化**: Redis 缓存用户会话、登录限流等热点数据
- 🔄 **接口解耦**: 通过接口定义避免循环依赖，提高代码可维护性
- 📝 **结构化日志**: 使用 Zap 实现高性能结构化日志记录
- ✅ **参数校验**: 集成 validator 并提供友好的中文错误提示

## 技术栈

| 类别 | 技术 |
|------|------|
| **Web 框架** | Gin |
| **数据库** | MySQL 8.0+ (GORM) |
| **缓存** | Redis 6.0+ |
| **认证** | JWT |
| **依赖注入** | Uber Fx |
| **日志** | Zap (结构化日志) |
| **ID 生成** | Snowflake 算法 |
| **参数校验** | go-playground/validator |
| **热重载** | Air |

## 目录结构

```text
uc-server/
├── cmd/
│   └── server/
│       └── main.go                    # 程序入口
│
├── config/
│   ├── config.go                      # 配置加载
│   └── config.yaml                    # 配置文件
│
├── internal/
│   ├── app/
│   │   └── app.go                     # Fx 应用初始化和生命周期管理
│   │
│   ├── middleware/                    # 中间件
│   │   ├── auth_handler.go           # JWT 认证中间件（接口解耦设计）
│   │   ├── body_size_limit.go        # 请求体大小限制
│   │   └── error_handler.go          # 统一错误处理
│   │
│   ├── model/                         # 数据模型
│   │   ├── base.go                    # 基础模型（ID、时间戳等）
│   │   ├── constants.go               # 常量定义
│   │   ├── user.go                    # 用户模型
│   │   ├── tenant.go                  # 租户模型
│   │   ├── tenant_user.go             # 租户-用户关联
│   │   ├── employee.go                # 员工模型
│   │   └── session.go                 # 会话数据模型
│   │
│   ├── modules/                       # 业务模块（DDD 模块化设计）
│   │   ├── modules.go                 # 模块汇总
│   │   │
│   │   ├── auth/                      # 认证模块
│   │   │   ├── auth.module.go        # Fx 模块定义
│   │   │   ├── auth.controller.go    # 认证控制器
│   │   │   ├── auth.service.go       # 认证服务（登录、登出、会话管理）
│   │   │   ├── auth.cache.go         # 认证缓存（Token 黑名单、限流）
│   │   │   └── auth.dto.go           # 数据传输对象
│   │   │
│   │   ├── user/                      # 用户模块
│   │   │   ├── user.module.go        # Fx 模块定义
│   │   │   ├── user.controller.go    # 用户控制器
│   │   │   ├── user.service.go       # 用户服务
│   │   │   ├── user.repository.go    # 用户数据访问
│   │   │   ├── user.cache.go         # 用户缓存（会话、登录限流）
│   │   │   └── user.dto.go           # 数据传输对象
│   │   │
│   │   └── health/                    # 健康检查模块
│   │       ├── health.module.go
│   │       └── health.controller.go
│   │
│   └── router/                        # 路由管理
│       ├── router.go                  # Gin Engine 初始化
│       └── api/
│           └── router.go              # API 路由注册
│
├── pkg/                               # 公共工具包
│   ├── errors/                        # 错误定义
│   │   ├── api_error.go
│   │   └── repository_error.go
│   │
│   ├── httpserver/                    # HTTP 服务器封装
│   │   ├── server.go
│   │   └── options.go
│   │
│   ├── jwt/                           # JWT 工具
│   │   ├── jwt.go
│   │   └── options.go
│   │
│   ├── logger/                        # 日志工具
│   │   └── logger.go
│   │
│   ├── mysql/                         # MySQL 连接封装
│   │   ├── mysql.go
│   │   └── options.go
│   │
│   ├── redis/                         # Redis 连接封装
│   │   ├── redis.go
│   │   └── options.go
│   │
│   ├── repository/                    # 通用仓储
│   │   └── base.go                    # 泛型 BaseRepository
│   │
│   ├── response/                      # 统一响应
│   │   ├── response.go
│   │   └── code.go
│   │
│   ├── snowflake/                     # 分布式 ID 生成
│   │   ├── snowflake.go
│   │   └── snowflake_test.go
│   │
│   ├── utils/                         # 工具函数
│   │   └── token.go
│   │
│   └── validation/                    # 参数校验
│       ├── validator_error.go         # 错误翻译（支持中文）
│       ├── validator_error_test.go
│       ├── password.go                # 密码强度校验
│       └── password_test.go
│
├── scripts/
│   └── init.sql                       # 数据库初始化脚本
│
├── deploy/
│   └── Dockerfile                     # Docker 镜像构建
│
├── .air.toml                          # Air 热重载配置
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 架构设计

### 模块化架构

项目采用 DDD（领域驱动设计）的模块化架构，每个业务模块包含完整的分层结构：

```
Module (auth/user/health)
├── module.go      # Fx 模块定义，提供依赖注入配置
├── controller.go  # HTTP 控制器，处理请求和响应
├── service.go     # 业务逻辑层，封装核心业务
├── repository.go  # 数据访问层，封装数据库操作
├── cache.go       # 缓存层，封装 Redis 操作
└── dto.go         # 数据传输对象，定义 API 接口
```

### 依赖注入流程

使用 Uber Fx 实现自动依赖注入和生命周期管理：

```
cmd/server/main.go
  └── app.Run(config)
        └── fx.New(
              fx.Supply(config),
              fx.Provide(
                provideLogger,
                provideMysql,
                provideRedis,
                provideJWT,
                router.NewEngine,
                api_router.NewAPIRouter,
                provideHTTPServer,
              ),
              modules.APIModule,  # 包含所有业务模块
              fx.Invoke(startHTTPServer),
            )

modules.APIModule
  ├── auth.Module
  │   ├── ProvideAuthCache      → *AuthCache + TokenBlacklistChecker 接口
  │   ├── ProvideAuthService    → AuthService + SessionDataProvider 接口
  │   └── NewAuthController     → *AuthController
  │
  ├── user.Module
  │   ├── NewUserCache          → *UserCache
  │   ├── NewUserRepository     → UserRepository 接口
  │   ├── NewTenantUserRepository → TenantUserRepository 接口
  │   ├── NewUserService        → UserService 接口
  │   └── NewUserController     → *UserController
  │
  └── health.Module
      └── NewHealthController   → *HealthController
```

### 接口解耦设计

为避免循环依赖，middleware 通过接口与业务模块解耦：

```go
// middleware 定义接口
type TokenBlacklistChecker interface {
    IsInBlacklistByTokenID(ctx context.Context, tokenString string) (bool, error)
}

type SessionDataProvider interface {
    GetUserSessionData(ctx *gin.Context, userID, tenantID uint64, claimsID string) (*model.UserSessionData, error)
}

// auth 模块实现接口
type AuthCache struct { ... }
func (c *AuthCache) IsInBlacklistByTokenID(...) (bool, error) { ... }

type authService struct { ... }
func (s *authService) GetUserSessionData(...) (*model.UserSessionData, error) { ... }

// Fx 自动注入
func ProvideAuthCache(redis *redis.Redis) AuthCacheResult {
    cache := NewAuthCache(redis)
    return AuthCacheResult{
        AuthCache:        cache,                    // 具体类型
        BlacklistChecker: cache,                    // 接口类型
    }
}
```

### 多租户设计

采用**共享数据库、共享 Schema** 模式：

- 通过 `tenant_id` 字段区分租户数据
- `tenant_user` 表实现用户与租户的多对多关联
- 支持一个用户属于多个租户，每个租户有不同角色
- 全局管理员（`super_admin`）可跨租户操作

### 缓存策略

Redis 缓存设计：

| 缓存类型 | Key 格式 | TTL | 用途 |
|---------|---------|-----|------|
| 用户信息 | `user:{id}` | 30分钟 | 减少数据库查询 |
| 用户会话 | `user:session:{uid}:{sid}` | 24小时 | JWT 会话数据 |
| Token 黑名单 | `auth:blacklist:{token_id}` | Token 剩余时间 | 登出后禁用 Token |
| 登录限流 | `user:login:attempts:{username}` | 1分钟 | 防暴力破解 |
| IP 限流 | `user:login:ip:attempts:{ip}` | 10分钟 | IP 级别限流 |

## API 接口

### 认证接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |
| POST | `/api/v1/auth/logout` | 用户登出 | ✅ |
| POST | `/api/v1/auth/change-password` | 修改密码 | ✅ |

### 用户接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/user/me` | 获取当前用户信息 | ✅ |
| GET | `/api/v1/user/list` | 获取用户列表 | ✅ |
| POST | `/api/v1/user/create` | 创建用户 | ✅ |
| GET | `/api/v1/user/info` | 获取用户详情 | ✅ |
| POST | `/api/v1/user/update` | 更新用户信息 | ✅ |
| POST | `/api/v1/user/delete` | 删除用户 | ✅ |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health/ready` | 就绪检查 |
| GET | `/health/live` | 存活检查 |

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

编辑 `config/config.yaml`：

```yaml
app:
  name: "uc-server"

server:
  port: 8090
  is_prod: false

database:
  mysql_url: "root:password@tcp(localhost:3306)/uc_server?charset=utf8mb4&parseTime=True&loc=Local"
  mysql_pool_max: 10

redis:
  url: "redis://localhost:6379/0"
  pool_max: 10

jwt:
  secret_key: "your-secret-key"
  expires_in: 86400  # 24小时

log:
  level: "info"
  dir: "./logs"
```

### 初始化数据库

```bash
mysql -u root -p < scripts/init.sql
```

### 运行项目

```bash
# 开发模式（带热重载）
air

# 或直接运行
go run cmd/server/main.go

# 编译生产版本
go build -o server cmd/server/main.go
./server
```

### 使用 Makefile

```bash
# 运行项目
make run

# 编译
make build

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint
```

## 开发规范

### 代码组织

- 每个业务模块独立管理，包含完整的分层结构
- 使用 Fx 模块化管理依赖注入
- 通过接口定义避免循环依赖

### 命名规范

- 文件名：小写 + 下划线（`user_service.go`）
- 包名：小写单词（`auth`, `user`）
- 接口名：大写开头 + 描述性名称（`UserRepository`, `SessionDataProvider`）
- 结构体：大写开头（`UserService`, `AuthCache`）

### 日志规范

使用结构化日志（key-value 格式）：

```go
// ✅ 推荐
logger.Infow("用户登录成功",
    "user_id", userID,
    "username", username,
    "client_ip", clientIP,
)

// ❌ 不推荐
logger.Infof("用户 %s 登录成功，IP: %s", username, clientIP)
```

### 错误处理

- 使用 `pkg/errors` 定义业务错误
- 使用 `pkg/response` 统一响应格式
- 参数校验错误自动翻译为中文

### 测试规范

- 单元测试文件以 `_test.go` 结尾
- 使用 table-driven tests 模式
- 关键业务逻辑必须有测试覆盖

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t uc-server:latest .

# 运行容器
docker run -d \
  -p 8090:8090 \
  -v $(pwd)/config:/app/config \
  --name uc-server \
  uc-server:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  uc-server:
    build: .
    ports:
      - "8090:8090"
    depends_on:
      - mysql
      - redis
    environment:
      - CONFIG_PATH=/app/config/config.yaml

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: uc_server
    ports:
      - "3306:3306"

  redis:
    image: redis:6.0
    ports:
      - "6379:6379"
```

## 性能优化

- ✅ Redis 缓存热点数据，减少数据库查询
- ✅ 使用连接池管理数据库和 Redis 连接
- ✅ 结构化日志提高日志性能
- ✅ JWT 无状态认证，减少服务器压力
- ✅ 泛型 BaseRepository 减少重复代码

## 安全特性

- 🔐 JWT Token 认证
- 🚫 Token 黑名单机制（登出后立即失效）
- 🔒 密码加密存储（bcrypt）
- 🛡️ 登录限流（防暴力破解）
- 🌐 IP 级别限流
- ✅ 参数校验防止注入攻击

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
