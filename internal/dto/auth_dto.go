package dto

import (
	"time"

	"github.com/yuhang1130/gin-server/internal/model"
)

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,omitempty,min=3,max=32,alphanum"`
	Email    string `json:"email" binding:"required,omitempty,email"`
	Password string `json:"password" binding:"required,omitempty,min=8,max=64"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,omitempty"` // 支持用户名或邮箱
	Password string `json:"password" binding:"required,omitempty"`
}

type Tenant struct {
	TenantID   uint64          `json:"tenant_id"`
	TenantName string          `json:"tenant_name"`
	Role       model.UserRoles `json:"role"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken          string        `json:"access_token"`
	AccessTokenExpiresAt time.Time     `json:"access_token_expires_at"`
	User                 *UserResponse `json:"user"`
	CurrentTenant        Tenant        `jsonL:"current_tenant"`
	Tenants              []Tenant      `json:"tenants"`
	IsGlobalAdmin        bool          `json:"is_global_admin"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}
