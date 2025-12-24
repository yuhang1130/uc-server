package config

import (
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port        string `mapstructure:"port"`
		Mode        string `mapstructure:"mode"`
		Timeout     int    `mapstructure:"timeout"`
		CorsAllowed string `mapstructure:"cors_allowed"`
		MaxBodySize int64  `mapstructure:"max_body_size"` // 最大请求体大小，单位字节
		IsDev       bool   `mapstructure:"is_dev"`
		IsQa        bool   `mapstructure:"is_qa"`
		IsProd      bool   `mapstructure:"is_prod"`
	} `mapstructure:"server"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		MaxConns int    `mapstructure:"max_conns"`
		MaxIdle  int    `mapstructure:"max_idle_conns"`
	} `mapstructure:"database"`

	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	JWT struct {
		SecretKey string `mapstructure:"secret_key"`
		ExpiresIn int    `mapstructure:"expires_in"`
	} `mapstructure:"jwt"`

	Snowflake struct {
		NodeID int64 `mapstructure:"node_id"` // 节点ID，范围 0-1023
	} `mapstructure:"snowflake"`

	Auth struct {
		RequireEmailVerification bool `mapstructure:"require_email_verification"` // 是否要求邮箱验证后才能登录
	} `mapstructure:"auth"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/app/config")
	viper.AddConfigPath("$HOME/.app/config")

	// 读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using environment variables")
		} else {
			return nil, err
		}
	}

	// 监听配置文件变化
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
	})
	viper.WatchConfig()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 设置默认值
	setDefaults(&config)
	// log.Printf("Parse Config Success. %+v \n", config)

	return &config, nil
}

func setDefaults(config *Config) {
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "dev"
	}
	if config.Server.MaxBodySize == 0 {
		config.Server.MaxBodySize = 10 * 1024 * 1024 // 默认 10MB
	}
	if config.Snowflake.NodeID == 0 {
		config.Snowflake.NodeID = 1 // 默认节点 ID 为 1
	}
	switch config.Server.Mode {
	case "debug", "dev":
		config.Server.IsDev = true
	case "qa":
		config.Server.IsQa = true
	case "production":
		config.Server.IsProd = true
	}
	if config.JWT.ExpiresIn == 0 {
		config.JWT.ExpiresIn = 7 * 24 * 3600 // 默认 7 天
	}
	// 其他默认值设置...
}
