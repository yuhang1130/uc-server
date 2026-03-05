package model

// UserResponse 用户响应（不包含敏感信息）
type UserResponse struct {
	ID        uint64     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email,omitempty"`
	Role      UserRoles  `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt int64      `json:"created_at"`
	UpdatedAt int64      `json:"updated_at"`
}

type TenantResponse struct {
	TenantID   uint64    `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Role       UserRoles `json:"role"`
}

// UserSessionData 用户会话数据结构，用于缓存用户登录会话
type UserSessionData struct {
	UserID        uint64           `json:"user_id"`
	User          *UserResponse    `json:"user"`
	TenantID      uint64           `json:"tenant_id"`
	TenantName    string           `json:"tenant_name"`
	Role          string           `json:"role"`
	Tenants       []TenantResponse `json:"tenants"`
	IsGlobalAdmin bool             `json:"is_global_admin"`
	IsProxy       bool             `json:"is_proxy"`      // 是否是代理token
	ProxyUserID   uint64           `json:"proxy_user_id"` // 被代理用户ID
}
