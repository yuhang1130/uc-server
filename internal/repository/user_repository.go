package repository

import (
	"errors"

	"github.com/yuhang1130/gin-server/internal/model"
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
	*BaseRepository
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
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
	// GORM 的 Delete 方法会自动使用软删除（如果模型包含 DeletedAt 字段）
	return r.DB.Where("id = ?", id).Delete(&model.User{}).Error
}

// HardDelete 物理删除用户（谨慎使用）
func (r *userRepository) HardDelete(id uint64) error {
	return r.DB.Unscoped().Delete(&model.User{}, id).Error
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
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 通过邮箱查找用户
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
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
