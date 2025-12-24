package service

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/internal/pkg/cache"
	"github.com/yuhang1130/gin-server/internal/repository"
	"go.uber.org/zap"
)

var validate = validator.New()

// UserService 用户服务接口
type UserService interface {
	CreateUser(ctx *gin.Context, user *model.User) error
	UpdateUser(ctx *gin.Context, id uint64, user *model.User) error
	DeleteUser(ctx *gin.Context, id uint64) error
	GetUserByID(ctx *gin.Context, id uint64) (*model.User, error)
	GetUserByEmail(ctx *gin.Context, email string) (*model.User, error)
	ListUsers(ctx *gin.Context, page, pageSize int) ([]*model.User, int, error)
	ListUsersWithFilter(ctx *gin.Context, filter *repository.UserListFilter) ([]*model.User, int, error)
}

// userService 用户服务实现
type userService struct {
	userRepo  repository.UserRepository
	userCache *cache.UserCache
	logger    *zap.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, userCache *cache.UserCache, logger *zap.Logger) UserService {
	return &userService{
		userRepo:  userRepo,
		userCache: userCache,
		logger:    logger,
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx *gin.Context, user *model.User) error {
	// 验证用户输入
	if err := validate.Struct(user); err != nil {
		s.logger.Error("Validation error: ", zap.Error(err))
		return errors.New("invalid user data")
	}

	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(user.Username); err == nil {
		return errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(user.Email); err == nil {
		return errors.New("email already exists")
	}

	// 设置默认角色
	if user.Role == "" {
		user.Role = "user"
	}

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user: ", zap.Error(err))
		return errors.New("failed to create user")
	}

	return nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx *gin.Context, id uint64, updatedUser *model.User) error {
	// 获取现有用户
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 检查用户名是否被更改且已存在
	if updatedUser.Username != "" && updatedUser.Username != user.Username {
		if _, err := s.userRepo.FindByUsername(updatedUser.Username); err == nil {
			return errors.New("username already exists")
		}
		user.Username = updatedUser.Username
	}

	// 检查邮箱是否被更改且已存在
	if updatedUser.Email != "" && updatedUser.Email != user.Email {
		if _, err := s.userRepo.FindByEmail(updatedUser.Email); err == nil {
			return errors.New("email already exists")
		}
		user.Email = updatedUser.Email
	}

	if updatedUser.Role != "" {
		user.Role = updatedUser.Role
	}
	// 这里简单地总是更新（实际应该从请求中判断是否提供了该字段）
	user.Status = updatedUser.Status

	// 更新用户
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user: ", zap.Error(err))
		return errors.New("failed to update user")
	}

	// 清除用户缓存，确保数据一致性
	if s.userCache != nil {
		bgCtx := context.Background()
		if err := s.userCache.DeleteUser(bgCtx, user.ID, user.Email); err != nil {
			s.logger.Warn("Failed to invalidate user cache after update", zap.Error(err), zap.Uint64("userID", user.ID))
		}
	}

	return nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx *gin.Context, id uint64) error {
	// 检查用户是否存在
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 删除用户
	if err := s.userRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete user: ", zap.Error(err))
		return errors.New("failed to delete user")
	}

	// 清除用户缓存
	if s.userCache != nil {
		bgCtx := context.Background()
		if err := s.userCache.DeleteUser(bgCtx, user.ID, user.Email); err != nil {
			s.logger.Warn("Failed to invalidate user cache after delete", zap.Error(err), zap.Uint64("userID", user.ID))
		}
	}

	return nil
}

// GetUserByID 通过ID获取用户
func (s *userService) GetUserByID(ctx *gin.Context, id uint64) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *userService) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers 分页获取用户列表
func (s *userService) ListUsers(ctx *gin.Context, page, pageSize int) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list users: ", zap.Error(err))
		return nil, 0, errors.New("failed to fetch users")
	}

	return users, total, nil
}

// ListUsersWithFilter 根据筛选条件分页获取用户列表
func (s *userService) ListUsersWithFilter(ctx *gin.Context, filter *repository.UserListFilter) ([]*model.User, int, error) {
	// 验证和设置默认值
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	users, total, err := s.userRepo.ListWithFilter(filter)
	if err != nil {
		s.logger.Error("Failed to list users with filter: ", zap.Error(err))
		return nil, 0, errors.New("failed to fetch users")
	}

	return users, total, nil
}
