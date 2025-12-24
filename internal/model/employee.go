package model

// Employee 员工表（存储组织内部的员工信息）
type Employee struct {
	BaseModel
	TenantID   uint64  `gorm:"type:int unsigned;not null;index:idx_tenant" json:"tenant_id" binding:"required"`
	UserID     *uint64 `gorm:"type:int unsigned;index" json:"user_id,omitempty"`
	Name       string  `gorm:"type:varchar(50);not null" json:"name" binding:"required"`
	Position   *string `gorm:"type:varchar(100)" json:"position,omitempty"`
	Department *string `gorm:"type:varchar(100)" json:"department,omitempty"`
	Email      *string `gorm:"type:varchar(100);index:idx_email,composite:deleted_at" json:"email,omitempty"`
	Status     int     `gorm:"type:tinyint;default:1" json:"status" binding:"required"`
}

// TableName 指定表名
func (Employee) TableName() string {
	return "employee"
}
