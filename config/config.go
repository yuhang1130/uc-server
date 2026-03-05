package config

import (
	"fmt"

	"github.com/spf13/viper"
)

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

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type ServerConfig struct {
	Port        string `mapstructure:"port"`
	Mode        string `mapstructure:"mode"`
	Timeout     int    `mapstructure:"timeout"`
	CorsAllowed string `mapstructure:"cors_allowed"`
	CookieName  string `mapstructure:"cookie_name"`
	MaxBodySize int64  `mapstructure:"max_body_size"`
	IsDev       bool   `mapstructure:"-"`
	IsQa        bool   `mapstructure:"-"`
	IsProd      bool   `mapstructure:"-"`
}

type DatabaseConfig struct {
	MysqlURL     string `mapstructure:"mysql_url"`
	MysqlPoolMax int    `mapstructure:"mysql_pool_max"`
}

type RedisConfig struct {
	URL     string `mapstructure:"url"`
	PoolMax int    `mapstructure:"pool_max"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

type LogConfig struct {
	Dir   string `mapstructure:"dir"`
	Level string `mapstructure:"level"`
}
type GoogleConfig struct {
	ClientID     string `env:"client_id"`
	ClientSecret string `env:"client_secret"`
	RedirectURL  string `env:"redirect_url"`
}

type GithubConfig struct {
	ClientID     string `env:"client_id"`
	ClientSecret string `env:"client_secret"`
	RedirectURL  string `env:"redirect_url"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.timeout", 30)
	viper.SetDefault("server.cors_allowed", "*")
	viper.SetDefault("jwt.expires_in", 604800)
	viper.SetDefault("log.dir", "./logs")
	viper.SetDefault("log.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
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
	if config.Log.Dir == "" {
		config.Log.Dir = "./logs"
	}
	if config.Log.Level == "" {
		config.Log.Level = "info"
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
}
