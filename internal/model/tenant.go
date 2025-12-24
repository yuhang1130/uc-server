package model

// Tenant 租户表（租户是企业）
type Tenant struct {
	BaseModel
	TenantName string `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_name,composite:deleted_at" json:"tenant_name" binding:"required"`
	Status     int    `gorm:"type:tinyint;default:1" json:"status" binding:"required"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenant"
}
