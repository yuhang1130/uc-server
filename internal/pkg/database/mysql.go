package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/yuhang1130/gin-server/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQL struct {
	DB    *gorm.DB
	SqlDB *sql.DB
}

func NewMySQL(cfg *config.Config) (*MySQL, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	log.Println("Connected to Mysql successfully.")

	// 连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &MySQL{DB: db, SqlDB: sqlDB}, nil
}

func (m *MySQL) Close() error {
	return m.SqlDB.Close()
}

// WithTransaction 创建一个事务上下文
func (m *MySQL) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
