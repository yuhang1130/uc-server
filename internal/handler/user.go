package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/internal/dto"
	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/internal/pkg/response"
	"github.com/yuhang1130/gin-server/internal/repository"
	"github.com/yuhang1130/gin-server/internal/service"
	"go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("创建用户请求参数无效",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	h.logger.Info("收到创建用户请求",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.String("client_ip", c.ClientIP()),
	)

	// 构建用户模型
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
		Status:   1,
	}

	// 如果没有指定角色，默认为 user
	if user.Role == "" {
		user.Role = "user"
	}

	// 设置密码
	if err := user.SetPassword(req.Password); err != nil {
		h.logger.Error("密码加密失败",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to hash password")
		return
	}

	// 创建用户
	if err := h.userService.CreateUser(c, user); err != nil {
		h.logger.Error("创建用户失败",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		response.BadRequest(c, "Failed to create user: "+err.Error())
		return
	}

	h.logger.Info("用户创建成功",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	// 构建响应
	userResp := &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.Success(c, userResp)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("获取用户列表请求参数无效",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	h.logger.Info("获取用户列表",
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize),
		zap.String("search", req.Search),
		zap.String("role", string(req.Role)),
		zap.String("client_ip", c.ClientIP()),
	)

	// 构建筛选条件
	filter := &repository.UserListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Search,
		Role:     req.Role,
		Status:   &req.Status,
	}

	// 获取用户列表（带筛选）
	users, total, err := h.userService.ListUsersWithFilter(c, filter)
	if err != nil {
		h.logger.Error("获取用户列表失败",
			zap.Int("page", req.Page),
			zap.Int("page_size", req.PageSize),
			zap.String("search", req.Search),
			zap.String("role", string(req.Role)),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to get user list: "+err.Error())
		return
	}

	h.logger.Info("用户列表获取成功",
		zap.Int("total", total),
		zap.Int("count", len(users)),
	)

	// 构建用户响应列表
	userResponses := make([]*dto.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, &dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// 计算总页数
	totalPages := (total + req.PageSize - 1) / req.PageSize

	response.Success(c, &dto.ListUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	var req dto.GetUserByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("获取用户详情请求参数无效",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	h.logger.Info("获取用户详情",
		zap.Uint64("user_id", req.ID),
		zap.String("client_ip", c.ClientIP()),
	)

	// 获取用户
	user, err := h.userService.GetUserByID(c, req.ID)
	if err != nil {
		h.logger.Warn("用户不存在",
			zap.Uint64("user_id", req.ID),
			zap.Error(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	h.logger.Info("用户详情获取成功",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// 构建响应
	userResp := &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.Success(c, userResp)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新用户请求参数无效",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	h.logger.Info("收到更新用户请求",
		zap.Uint64("user_id", req.ID),
		zap.String("client_ip", c.ClientIP()),
	)

	// 获取现有用户
	existingUser, err := h.userService.GetUserByID(c, req.ID)
	if err != nil {
		h.logger.Warn("更新用户失败：用户不存在",
			zap.Uint64("user_id", req.ID),
			zap.Error(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	// 只更新提供的字段
	if req.Username != nil {
		existingUser.Username = *req.Username
	}
	if req.Email != nil {
		existingUser.Email = *req.Email
	}
	if req.Role != nil {
		existingUser.Role = *req.Role
	}
	if req.Status != nil {
		existingUser.Status = *req.Status
	}

	// 更新用户
	if err := h.userService.UpdateUser(c, req.ID, existingUser); err != nil {
		h.logger.Error("更新用户失败",
			zap.Uint64("user_id", req.ID),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to update user: "+err.Error())
		return
	}

	h.logger.Info("用户更新成功",
		zap.Uint64("user_id", existingUser.ID),
		zap.String("username", existingUser.Username),
	)

	// 构建响应
	userResp := &dto.UserResponse{
		ID:        existingUser.ID,
		Username:  existingUser.Username,
		Email:     existingUser.Email,
		Role:      existingUser.Role,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
	}

	response.Success(c, userResp)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req dto.DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("删除用户请求参数无效",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	h.logger.Info("收到删除用户请求",
		zap.Uint64("user_id", req.ID),
		zap.String("client_ip", c.ClientIP()),
	)

	// 删除用户
	if err := h.userService.DeleteUser(c, req.ID); err != nil {
		if err.Error() == "user not found" {
			h.logger.Warn("删除用户失败：用户不存在",
				zap.Uint64("user_id", req.ID),
			)
			response.NotFound(c, err.Error())
		} else {
			h.logger.Error("删除用户失败",
				zap.Uint64("user_id", req.ID),
				zap.Error(err),
			)
			response.InternalServerError(c, err.Error())
		}
		return
	}

	h.logger.Info("用户删除成功",
		zap.Uint64("user_id", req.ID),
	)

	response.Success(c, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// 从上下文中获取当前用户
	user, exists := c.Get("currentUser")
	if !exists {
		h.logger.Warn("获取当前用户失败：未认证",
			zap.String("client_ip", c.ClientIP()),
		)
		response.Unauthorized(c, "User not authenticated")
		return
	}

	currentUser, ok := user.(*model.User)
	if !ok {
		h.logger.Error("获取当前用户失败：类型断言失败",
			zap.String("client_ip", c.ClientIP()),
		)
		response.InternalServerError(c, "Failed to get current user")
		return
	}

	h.logger.Info("获取当前用户成功",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
	)

	// 构建响应
	userResp := &dto.UserResponse{
		ID:        currentUser.ID,
		Username:  currentUser.Username,
		Email:     currentUser.Email,
		Role:      currentUser.Role,
		CreatedAt: currentUser.CreatedAt,
		UpdatedAt: currentUser.UpdatedAt,
	}

	response.Success(c, userResp)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	// 从上下文中获取当前用户
	user, exists := c.Get("currentUser")
	if !exists {
		h.logger.Warn("更新密码失败：未认证",
			zap.String("client_ip", c.ClientIP()),
		)
		response.Unauthorized(c, "User not authenticated")
		return
	}

	currentUser, ok := user.(*model.User)
	if !ok {
		h.logger.Error("更新密码失败：类型断言失败",
			zap.String("client_ip", c.ClientIP()),
		)
		response.InternalServerError(c, "Failed to get current user")
		return
	}

	// 绑定请求参数
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新密码请求参数无效",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	h.logger.Info("收到更新密码请求",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 验证旧密码
	if !currentUser.CheckPassword(req.OldPassword) {
		h.logger.Warn("更新密码失败：旧密码错误",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.String("client_ip", c.ClientIP()),
		)
		response.BadRequest(c, "Old password is incorrect")
		return
	}

	// 设置新密码
	if err := currentUser.SetPassword(req.NewPassword); err != nil {
		h.logger.Error("新密码加密失败",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to hash new password")
		return
	}

	// 更新用户
	if err := h.userService.UpdateUser(c, currentUser.ID, currentUser); err != nil {
		h.logger.Error("更新密码失败",
			zap.Uint64("user_id", currentUser.ID),
			zap.String("username", currentUser.Username),
			zap.Error(err),
		)
		response.InternalServerError(c, "Failed to update password")
		return
	}

	h.logger.Info("密码更新成功",
		zap.Uint64("user_id", currentUser.ID),
		zap.String("username", currentUser.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	response.Success(c, gin.H{"message": "Password updated successfully"})
}
