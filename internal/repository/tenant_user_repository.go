package repository

import (
	"errors"

	"github.com/yuhang1130/gin-server/internal/model"
	"gorm.io/gorm"
)

// TenantUserRepository 租户用户关联仓库接口
type TenantUserRepository interface {
	FindUserTenants(userID uint64) ([]*model.TenantUser, error)
	FindUserTenantsWithTenant(userID uint64) ([]*TenantUserWithTenant, error)
	UpdateLastLoginAt(tenantID, userID uint64) error
}

// TenantUserWithTenant 租户用户关联信息（包含租户详情）
type TenantUserWithTenant struct {
	model.TenantUser
	Tenant model.Tenant `gorm:"foreignKey:TenantID"`
}

// tenantUserRepository 租户用户关联仓库实现
type tenantUserRepository struct {
	*BaseRepository
}

// NewTenantUserRepository 创建租户用户关联仓库实例
func NewTenantUserRepository(db *gorm.DB) TenantUserRepository {
	return &tenantUserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// FindUserTenants 查找用户所属的所有租户
func (r *tenantUserRepository) FindUserTenants(userID uint64) ([]*model.TenantUser, error) {
	var tenantUsers []*model.TenantUser
	if err := r.DB.Where("user_id = ? AND status = ?", userID, 1).Find(&tenantUsers).Error; err != nil {
		return nil, err
	}
	return tenantUsers, nil
}

// FindUserTenantsWithTenant 查找用户所属的所有租户（包含租户详情）
func (r *tenantUserRepository) FindUserTenantsWithTenant(userID uint64) ([]*TenantUserWithTenant, error) {
	var results []*TenantUserWithTenant
	err := r.DB.Table("tenant_user").
		Select("tenant_user.*, tenant.*").
		Joins("LEFT JOIN tenant ON tenant_user.tenant_id = tenant.id").
		Where("tenant_user.user_id = ? AND tenant_user.status = ? AND tenant.status = ?", userID, 1, 1).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("user has no active tenants")
	}

	return results, nil
}

// UpdateLastLoginAt 更新用户在租户中的最后登录时间
func (r *tenantUserRepository) UpdateLastLoginAt(tenantID, userID uint64) error {
	return r.DB.Model(&model.TenantUser{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}
