package mysql

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second

	_logConnAttempts = "MySQL is trying to connect, attempts left: %d, err: %v"
)

// MySQL -.
type MySQL struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	DB *gorm.DB
}

// New -.
func New(dsn string, opts ...Option) (*MySQL, error) {
	m := &MySQL{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(m)
	}

	var err error

	for m.connAttempts > 0 {
		m.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Printf(_logConnAttempts, m.connAttempts, err)
			time.Sleep(m.connTimeout)
			m.connAttempts--
			continue
		}

		sqlDB, sqlErr := m.DB.DB()
		if sqlErr != nil {
			err = sqlErr
			log.Printf(_logConnAttempts, m.connAttempts, err)
			time.Sleep(m.connTimeout)
			m.connAttempts--
			continue
		}

		sqlDB.SetMaxIdleConns(m.maxPoolSize)
		sqlDB.SetMaxOpenConns(m.maxPoolSize)
		sqlDB.SetConnMaxLifetime(time.Hour)

		if pingErr := sqlDB.Ping(); pingErr != nil {
			err = pingErr
			log.Printf(_logConnAttempts, m.connAttempts, err)
			time.Sleep(m.connTimeout)
			m.connAttempts--
			continue
		}

		break
	}

	if err != nil {
		return nil, fmt.Errorf("mysql - New - connAttempts == 0: %w", err)
	}

	return m, nil
}

// Close -.
func (m *MySQL) Close() error {
	if m.DB != nil {
		sqlDB, err := m.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
