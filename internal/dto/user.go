package dto

import (
	"github.com/yuhang1130/gin-server/internal/model"
)

// UserResponse 用户响应（不包含敏感信息）
type UserResponse struct {
	ID        uint64           `json:"id"`
	Username  string           `json:"username"`
	Email     string           `json:"email,omitempty"`
	Role      model.UserRoles  `json:"role"`
	Status    model.UserStatus `json:"status"`
	CreatedAt int64            `json:"created_at"`
	UpdatedAt int64            `json:"updated_at"`
}

// CreateUserRequest 创建用户请求（管理员创建用户）
type CreateUserRequest struct {
	Username string          `json:"username" binding:"required,min=3,max=32,alphanum"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8,max=64"`
	Role     model.UserRoles `json:"role" binding:"omitempty,oneof=user tenant_admin super_admin"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID       uint64            `json:"id" binding:"required,min=1"`
	Username *string           `json:"username" binding:"omitempty,min=3,max=32,alphanum"`
	Email    *string           `json:"email" binding:"omitempty,email"`
	Role     *model.UserRoles  `json:"role" binding:"omitempty,oneof=user tenant_admin super_admin"`
	Status   *model.UserStatus `json:"status"`
}

// GetUserByIDRequest 根据ID获取用户请求
type GetUserByIDRequest struct {
	ID uint64 `form:"id" binding:"required,min=1"`
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	ID uint64 `form:"id" binding:"required,min=1"`
}

// ListUsersRequest 获取用户列表请求
type ListUsersRequest struct {
	Page     int              `form:"page" binding:"omitempty,min=1"`
	PageSize int              `form:"page_size" binding:"omitempty,min=1,max=1000"`
	Search   string           `form:"search" binding:"omitempty,max=100"`
	Role     model.UserRoles  `form:"role" binding:"omitempty,oneof=user tenant_admin super_admin"`
	Status   model.UserStatus `form:"status" binding:"omitempty,oneof=0 1"`
}

// ListUsersResponse 获取用户列表响应
type ListUsersResponse struct {
	Users      []*UserResponse `json:"users"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// UpdatePasswordRequest 更新密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}
