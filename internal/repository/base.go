package repository

import "gorm.io/gorm"

// BaseRepository 基础仓库
type BaseRepository struct {
	DB *gorm.DB
}

// NewBaseRepository 创建基础仓库实例
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// WithTrx 执行事务操作
func (r *BaseRepository) WithTrx(fn func(tx *gorm.DB) error) error {
	tx := r.DB.Begin()
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
