package model

// TenantUser 租户与用户关联表（表示用户属于哪个租户）
type TenantUser struct {
	BaseModel
	TenantID    uint64 `gorm:"primaryKey;type:int unsigned;not null" json:"tenant_id" binding:"required"`
	UserID      uint64 `gorm:"primaryKey;type:int unsigned;not null;index:idx_user" json:"user_id" binding:"required"`
	Role        string `gorm:"type:varchar(50);not null;index:idx_role" json:"role" binding:"required"`
	Status      int    `gorm:"type:tinyint;default:1" json:"status" binding:"required"`
	LastLoginAt int64  `json:"last_login_at"`
}

// TableName 指定表名
func (TenantUser) TableName() string {
	return "tenant_user"
}
