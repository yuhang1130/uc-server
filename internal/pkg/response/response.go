package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessFunc(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: "success",
		Data:    data,
	})
}

func ErrorFunc(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func BadRequestFunc(c *gin.Context, message string) {
	ErrorFunc(c, BadRequest, message)
}

func UnauthorizedFunc(c *gin.Context, message string) {
	ErrorFunc(c, Unauthorized, message)
}

func ForbiddenFunc(c *gin.Context, message string) {
	ErrorFunc(c, Forbidden, message)
}

func NotFoundFunc(c *gin.Context, message string) {
	ErrorFunc(c, NotFound, message)
}

func InternalServerErrorFunc(c *gin.Context, message string) {
	ErrorFunc(c, InternalServerError, message)
}

func TooManyRequestsFunc(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    TooManyRequests,
		Message: message,
	})
}

// 业务相关错误方法
func UserNotFoundErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "用户不存在"
	}
	ErrorFunc(c, UserNotFound, message)
}

func UserAlreadyExistsErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "用户已存在"
	}
	ErrorFunc(c, UserAlreadyExists, message)
}

func UserPasswordIncorrectErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "用户名或密码错误"
	}
	ErrorFunc(c, UserPasswordIncorrect, message)
}

func AuthInvalidCredentialsErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "无效的凭据"
	}
	ErrorFunc(c, AuthInvalidCredentials, message)
}

func ValidationErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "参数验证失败"
	}
	ErrorFunc(c, ValidationError, message)
}

func TenantNotFoundErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "租户不存在"
	}
	ErrorFunc(c, TenantNotFound, message)
}

func RecordNotFoundErrorFunc(c *gin.Context, message string) {
	if message == "" {
		message = "记录未找到"
	}
	ErrorFunc(c, DatabaseRecordNotFound, message)
}

func UnknownErrorFunc(c *gin.Context, message string) {
	ErrorFunc(c, UnknownError, message)
}

func UserTenantMismatchFunc(c *gin.Context, message string) {
	ErrorFunc(c, UserTenantMismatch, message)
}

func AuthLoginFailedFunc(c *gin.Context, message string) {
	ErrorFunc(c, AuthLoginFailed, message)
}

func CacheOperationFailedFunc(c *gin.Context, message string) {
	ErrorFunc(c, CacheOperationFailed, message)
}

func UserPasswordIncorrectFunc(c *gin.Context, message string) {
	ErrorFunc(c, UserPasswordIncorrect, message)
}

func DBOperationFailedFunc(c *gin.Context, message string) {
	ErrorFunc(c, DatabaseOperationFailed, message)
}

func SessionGetFailedFunc(c *gin.Context, message string) {
	ErrorFunc(c, CacheKeyNotFound, message)
}
