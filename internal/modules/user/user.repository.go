package user

import (
	"errors"

	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/pkg/mysql"
	"github.com/yuhang1130/gin-server/pkg/repository"
	"gorm.io/gorm"
)

// UserListFilter 用户列表筛选条件
type UserListFilter struct {
	Page     int
	PageSize int
	Search   string            // 搜索关键词（用户名或邮箱）
	Role     model.UserRoles   // 角色筛选
	Status   *model.UserStatus // 状态筛选
}

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id uint64) error
	FindByID(id uint64) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	List(page, pageSize int) ([]*model.User, int, error)
	ListWithFilter(filter *UserListFilter) ([]*model.User, int, error)
	Count() (int, error)
}

// userRepository 用户仓库实现
type userRepository struct {
	*repository.BaseRepository
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *mysql.MySQL) UserRepository {
	return &userRepository{
		BaseRepository: repository.NewBaseRepository(db.DB),
	}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.DB.Save(user).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uint64) error {
	return r.DB.Where("id = ?", id).Delete(&model.User{}).Error
}

// FindByID 通过ID查找用户
func (r *userRepository) FindByID(id uint64) (*model.User, error) {
	var user model.User
	if err := r.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 通过用户名查找用户
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 通过邮箱查找用户
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List 分页获取用户列表
func (r *userRepository) List(page, pageSize int) ([]*model.User, int, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * pageSize
	if err := r.DB.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.DB.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

// ListWithFilter 根据筛选条件分页获取用户列表
func (r *userRepository) ListWithFilter(filter *UserListFilter) ([]*model.User, int, error) {
	var users []*model.User
	var total int64

	// 构建查询
	query := r.DB.Model(&model.User{})

	// 应用搜索条件（搜索用户名或邮箱）
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("username LIKE ? OR email LIKE ?", searchPattern, searchPattern)
	}

	// 应用角色筛选
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	// 应用状态筛选
	if filter.Status != nil {
		query = query.Where("status = ?", filter.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

// Count 获取用户总数
func (r *userRepository) Count() (int, error) {
	var count int64
	if err := r.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

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
	*repository.BaseRepository
}

// NewTenantUserRepository 创建租户用户关联仓库实例
func NewTenantUserRepository(db *mysql.MySQL) TenantUserRepository {
	return &tenantUserRepository{
		BaseRepository: repository.NewBaseRepository(db.DB),
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
		Update("last_login_at", gorm.Expr("UNIX_TIMESTAMP(NOW()) * 1000")).Error
}
