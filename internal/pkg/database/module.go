package database

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func provideGormDB(m *MySQL) *gorm.DB {
	return m.DB
}

var Module = fx.Options(
	fx.Provide(NewMySQL),
	fx.Provide(provideGormDB),
)
