package model

import (
	"time"

	"github.com/yuhang1130/gin-server/internal/pkg/snowflake"
	"gorm.io/gorm"
)

// BaseModel 通用基础模型
type BaseModel struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook before creating record
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

// BeforeUpdate hook before updating record
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// Generate 生成唯一的 uint64 ID（使用雪花算法）
func Generate() uint64 {
	return snowflake.GenerateUint()
}

type TenantBaseModel struct {
	IDBaseModel
	TenantID *uint64 `gorm:"index;type:int unsigned" json:"tenant_id,omitempty"`
}

type IDBaseModel struct {
	BaseModel
	ID uint64 `gorm:"primaryKey" json:"id"`
}

// BeforeCreate hook before creating record with ID
func (m *IDBaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.ID == 0 {
		m.ID = Generate()
	}
	return nil
}
