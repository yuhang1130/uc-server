// Package config 提供应用配置的加载与管理。
// 支持从 YAML 配置文件或命令行指定路径读取，并自动绑定环境变量。
//
// 使用方式：
//
//	cfg, err := config.Load()
//	if err != nil { ... }
//
// 或通过单例获取：
//
//	cfg := config.Get()
package config

import (
	"flag"
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// ============================================================
// 环境类型定义
// ============================================================

// Environment 定义应用运行环境类型。
type Environment string

const (
	EnvDev  Environment = "dev"  // 开发环境
	EnvQa   Environment = "qa"   // 测试环境
	EnvProd Environment = "prod" // 生产环境
)

// IsDev 判断当前是否为开发环境。
func (e Environment) IsDev() bool { return e == EnvDev }

// IsQa 判断当前是否为测试环境。
func (e Environment) IsQa() bool { return e == EnvQa }

// IsProd 判断当前是否为生产环境。
func (e Environment) IsProd() bool { return e == EnvProd }

// ============================================================
// 配置结构体定义
// ============================================================

// Config 应用配置根结构，所有配置项通过 mapstructure 与 YAML 字段映射。
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Google   GoogleConfig   `mapstructure:"google"`
	Github   GithubConfig   `mapstructure:"github"`
}

// AppConfig 应用基础信息。
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// ServerConfig HTTP 服务相关配置。
type ServerConfig struct {
	Port        string      `mapstructure:"port"`          // 监听端口
	Mode        Environment `mapstructure:"mode"`          // 运行模式: dev|qa|prod
	Timeout     int         `mapstructure:"timeout"`       // 请求超时时间（秒）
	CorsAllowed string      `mapstructure:"cors_allowed"`  // CORS 允许的源
	CookieName  string      `mapstructure:"cookie_name"`   // Cookie 名称
	MaxBodySize int64       `mapstructure:"max_body_size"` // 最大请求体大小（字节）
}

// DatabaseConfig 数据库连接配置。
type DatabaseConfig struct {
	MysqlURL     string `mapstructure:"mysql_url"`      // MySQL 连接字符串
	MysqlPoolMax int    `mapstructure:"mysql_pool_max"` // 连接池最大连接数
}

// RedisConfig Redis 缓存配置。
type RedisConfig struct {
	URL     string `mapstructure:"url"`      // Redis 连接地址
	PoolMax int    `mapstructure:"pool_max"` // 连接池最大连接数
}

// JWTConfig JWT 认证配置。
type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"` // JWT 签名密钥
	ExpiresIn int    `mapstructure:"expires_in"` // Token 过期时间（秒）
}

// LogConfig 日志配置。
type LogConfig struct {
	Dir   string `mapstructure:"dir"`   // 日志输出目录
	Level string `mapstructure:"level"` // 日志级别: debug|info|warn|error
}

// GoogleConfig Google OAuth 配置。
type GoogleConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// GithubConfig Github OAuth 配置。
type GithubConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// ============================================================
// 全局单例
// ============================================================

var (
	globalCfg  *Config   // 全局配置实例
	globalOnce sync.Once // 确保 Load 只执行一次
	globalErr  error     // 首次加载的错误（若有）
)

// ============================================================
// 配置加载
// ============================================================

// Load 加载应用配置，首次调用会解析命令行参数并读取配置文件。
// 该方法线程安全，多次调用返回同一实例。
//
// 支持的命令行参数：
//
//	-env string    指定运行环境，对应 config/{env}.yaml（默认 "dev"）
//	-f string      直接指定配置文件路径，优先级高于 -env
//
// 使用示例：
//
//	go run cmd/server/main.go -env prod
//	go run cmd/server/main.go -f ./config/custom.yaml
func Load() (*Config, error) {
	globalOnce.Do(func() {
		globalCfg, globalErr = loadOnce()
	})
	return globalCfg, globalErr
}

// Get 返回已加载的全局配置实例。
// 注意：必须先调用 Load() 且成功返回后，Get() 才不会返回 nil。
func Get() *Config {
	return globalCfg
}

// loadOnce 执行实际的配置加载逻辑（内部使用）。
func loadOnce() (*Config, error) {
	// 定义命令行参数（仅在首次调用时生效）
	var (
		env        = flag.String("env", "dev", "config env: dev|qa|prod")
		configFile = flag.String("f", "", "the config file (optional, overrides -env)")
	)
	flag.Parse()

	// 配置 Viper 读取源
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
	} else {
		viper.SetConfigName(*env)
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}

	// 设置默认值（唯一来源，避免多处维护导致不一致）
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "dev")
	viper.SetDefault("server.timeout", 30)
	viper.SetDefault("server.cors_allowed", "*")
	viper.SetDefault("server.max_body_size", 10*1024*1024) // 10MB
	viper.SetDefault("jwt.expires_in", 7*24*3600)          // 7天
	viper.SetDefault("log.dir", "./logs")
	viper.SetDefault("log.level", "info")

	// 启用环境变量自动绑定（如 SERVER_PORT 会覆盖 server.port）
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// 反序列化到结构体
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 打印加载信息（使用 fmt，因为此时 zap 可能尚未初始化）
	fmt.Printf("[config] loaded: env=%s, file=%s\n", *env, viper.ConfigFileUsed())

	return &cfg, nil
}
