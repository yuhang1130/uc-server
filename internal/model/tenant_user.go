package model

import (
	"time"

	"gorm.io/gorm"
)

// TenantUser 租户与用户关联表（表示用户属于哪个租户）
type TenantUser struct {
	BaseModel
	TenantID    uint64    `gorm:"primaryKey;type:int unsigned;not null" json:"tenant_id" binding:"required"`
	UserID      uint64    `gorm:"primaryKey;type:int unsigned;not null;index:idx_user" json:"user_id" binding:"required"`
	Role        string    `gorm:"type:varchar(50);not null;index:idx_role" json:"role" binding:"required"`
	Status      int       `gorm:"type:tinyint;default:1" json:"status" binding:"required"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// TableName 指定表名
func (TenantUser) TableName() string {
	return "tenant_user"
}

// BeforeCreate hook before creating record
func (m *TenantUser) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

// BeforeUpdate hook before updating record
func (m *TenantUser) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
