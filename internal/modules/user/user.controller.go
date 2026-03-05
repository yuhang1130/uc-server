package user

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yuhang1130/gin-server/internal/model"
	"github.com/yuhang1130/gin-server/pkg/logger"
	"github.com/yuhang1130/gin-server/pkg/response"
	"github.com/yuhang1130/gin-server/pkg/validation"
)

const invalidRequestPrefix = "Invalid request: "

// UserController 用户控制器
type UserController struct {
	userService UserService
	logger      logger.Logger
}

// NewUserController 创建用户控制器实例
func NewUserController(userService UserService, logger logger.Logger) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		h.logger.Warnw("创建用户请求参数无效",
			"error", errMessage,
			"client_ip", c.ClientIP(),
		)
		response.ValidationErrorFunc(c, invalidRequestPrefix+errMessage)
		return
	}

	h.logger.Infow("收到创建用户请求",
		"username", req.Username,
		"email", req.Email,
		"client_ip", c.ClientIP(),
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
		h.logger.Errorw("密码加密失败",
			"username", req.Username,
			"error", err.Error(),
		)
		response.InternalServerErrorFunc(c, "Failed to hash password")
		return
	}

	// 创建用户
	if err := h.userService.CreateUser(c, user); err != nil {
		// 根据错误类型返回不同的状态码
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "exists") {
			response.UserAlreadyExistsErrorFunc(c, "Failed to create user: "+err.Error())
		} else {
			response.DBOperationFailedFunc(c, "Failed to create user: "+err.Error())
		}
		return
	}

	h.logger.Infow("用户创建成功",
		"user_id", user.ID,
		"username", user.Username,
		"email", user.Email,
	)

	// 构建响应
	userResp := &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.SuccessFunc(c, userResp)
}

func (h *UserController) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		h.logger.Warnw("获取用户列表请求参数无效",
			"error", errMessage,
			"client_ip", c.ClientIP(),
		)
		response.BadRequestFunc(c, invalidRequestPrefix+errMessage)
		return
	}

	h.logger.Infow("收到获取用户列表请求",
		"page", req.Page,
		"page_size", req.PageSize,
		"search", req.Search,
		"role", req.Role,
		"client_ip", c.ClientIP(),
	)

	// 设置分页默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 构建筛选条件
	filter := &UserListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Search,
		Role:     req.Role,
		Status:   &req.Status,
	}

	// 获取用户列表（带筛选）
	users, total, err := h.userService.ListUsersWithFilter(c, filter)
	if err != nil {
		response.InternalServerErrorFunc(c, "Failed to get user list: "+err.Error())
		return
	}

	h.logger.Infow("用户列表获取成功",
		"total", total,
		"count", len(users),
	)

	// 构建用户响应列表
	userResponses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, &UserResponse{
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

	response.SuccessFunc(c, &ListUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

func (h *UserController) GetUserByID(c *gin.Context) {
	var req GetUserByIDRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		h.logger.Warnw("获取用户详情请求参数无效",
			"error", errMessage,
			"client_ip", c.ClientIP(),
		)
		response.ValidationErrorFunc(c, invalidRequestPrefix+errMessage)
		return
	}

	h.logger.Infow("获取用户详情",
		"user_id", req.ID,
		"client_ip", c.ClientIP(),
	)

	// 获取用户
	user, err := h.userService.GetUserByID(c, req.ID)
	if err != nil {
		response.UserNotFoundErrorFunc(c, "User not found")
		return
	}

	h.logger.Infow("用户详情获取成功",
		"user_id", user.ID,
		"username", user.Username,
	)

	// 构建响应
	userResp := &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.SuccessFunc(c, userResp)
}

func (h *UserController) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		h.logger.Warnw("更新用户请求参数无效",
			"error", errMessage,
			"client_ip", c.ClientIP(),
		)
		response.ValidationErrorFunc(c, invalidRequestPrefix+errMessage)
		return
	}

	h.logger.Infow("收到更新用户请求",
		"user_id", req.ID,
		"client_ip", c.ClientIP(),
	)

	// 获取现有用户
	existingUser, err := h.userService.GetUserByID(c, req.ID)
	if err != nil {
		response.UserNotFoundErrorFunc(c, "User not found")
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
		response.DBOperationFailedFunc(c, "Failed to update user: "+err.Error())
		return
	}

	h.logger.Infow("用户更新成功",
		"user_id", existingUser.ID,
		"username", existingUser.Username,
	)

	// 构建响应
	userResp := &UserResponse{
		ID:        existingUser.ID,
		Username:  existingUser.Username,
		Email:     existingUser.Email,
		Role:      existingUser.Role,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
	}

	response.SuccessFunc(c, userResp)
}

func (h *UserController) DeleteUser(c *gin.Context) {
	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessage := validation.TranslateValidationError(err)
		h.logger.Warnw("删除用户请求参数无效",
			"error", errMessage,
			"client_ip", c.ClientIP(),
		)
		response.ValidationErrorFunc(c, invalidRequestPrefix+errMessage)
		return
	}

	h.logger.Infow("收到删除用户请求",
		"user_id", req.ID,
		"client_ip", c.ClientIP(),
	)

	// 删除用户
	if err := h.userService.DeleteUser(c, req.ID); err != nil {
		if err.Error() == "user not found" {
			response.UserNotFoundErrorFunc(c, err.Error())
		} else {
			response.DBOperationFailedFunc(c, err.Error())
		}
		return
	}

	h.logger.Infow("用户删除成功",
		"user_id", req.ID,
	)

	response.SuccessFunc(c, gin.H{"message": "User deleted successfully"})
}

func (h *UserController) GetCurrentUser(c *gin.Context) {
	// 从上下文中获取当前用户
	userSessionData, exists := c.Get("userSessionData")
	if !exists {
		h.logger.Warnw("获取当前用户失败：未认证",
			"client_ip", c.ClientIP(),
		)
		response.UnauthorizedFunc(c, "User not authenticated")
		return
	}

	currentUser, ok := userSessionData.(*model.UserSessionData)
	if !ok {
		h.logger.Errorw("获取当前用户失败：类型断言失败",
			"client_ip", c.ClientIP(),
		)
		response.InternalServerErrorFunc(c, "Failed to get current user")
		return
	}

	h.logger.Infow("获取当前用户成功",
		"user_id", currentUser.UserID,
		"username", currentUser.User.Username,
	)

	response.SuccessFunc(c, currentUser.User)
}
