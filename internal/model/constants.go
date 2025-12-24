package model

// UserStatus 用户状态类型
type UserStatus uint8

const (
	UserStatusInactive UserStatus = 0 // 未激活/禁用
	UserStatusActive   UserStatus = 1 // 激活/启用
)

// UserRole 用户角色类型
type UserRoles string

const (
	UserRoleUser        UserRoles = "user"         // 普通用户
	UserRoleAdminTenant UserRoles = "tenant_admin" // 租户管理员
	UserRoleAdminSystem UserRoles = "super_admin"  // 系统管理员
)
