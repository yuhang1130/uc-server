package auth

import (
	"github.com/yuhang1130/gin-server/internal/model"
)

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 支持用户名或邮箱
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken          string                 `json:"access_token"`
	AccessTokenExpiresAt int64                  `json:"access_token_expires_at"`
	User                 *model.UserResponse    `json:"user"`
	CurrentTenant        model.TenantResponse   `json:"current_tenant"`
	Tenants              []model.TenantResponse `json:"tenants"`
	IsGlobalAdmin        bool                   `json:"is_global_admin"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint64           `json:"id"`
	Username  string           `json:"username"`
	Email     string           `json:"email"`
	Role      model.UserRoles  `json:"role"`
	Status    model.UserStatus `json:"status"`
	CreatedAt int64            `json:"created_at"`
	UpdatedAt int64            `json:"updated_at"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}
