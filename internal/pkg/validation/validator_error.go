package validation

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
)

// TranslateValidationError 将验证错误翻译为友好的中文消息
func TranslateValidationError(err error) string {
	// 检查是否为空请求体错误
	if errors.Is(err, io.EOF) {
		return "请求体不能为空"
	}

	// 检查是否为验证错误
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		// 收集所有字段的错误信息
		var errMessages []string
		for _, e := range validationErrors {
			errMessages = append(errMessages, translateFieldError(e))
		}
		return strings.Join(errMessages, "; ")
	}

	// 其他错误直接返回
	return err.Error()
}

// translateFieldError 翻译单个字段的验证错误
func translateFieldError(fe validator.FieldError) string {
	field := translateFieldName(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "email":
		return fmt.Sprintf("%s格式不正确", field)
	case "min":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s长度不能少于%s个字符", field, fe.Param())
		}
		return fmt.Sprintf("%s不能小于%s", field, fe.Param())
	case "max":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s长度不能超过%s个字符", field, fe.Param())
		}
		return fmt.Sprintf("%s不能大于%s", field, fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s只能包含字母和数字", field)
	case "numeric":
		return fmt.Sprintf("%s必须是数字", field)
	case "alpha":
		return fmt.Sprintf("%s只能包含字母", field)
	case "len":
		return fmt.Sprintf("%s长度必须为%s", field, fe.Param())
	case "eq":
		return fmt.Sprintf("%s必须等于%s", field, fe.Param())
	case "ne":
		return fmt.Sprintf("%s不能等于%s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s必须大于%s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s必须大于或等于%s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s必须小于%s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s必须小于或等于%s", field, fe.Param())
	case "url":
		return fmt.Sprintf("%s必须是有效的URL", field)
	case "uri":
		return fmt.Sprintf("%s必须是有效的URI", field)
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", field, fe.Param())
	case "is_mobile":
		return fmt.Sprintf("%s格式不正确", field)
	default:
		// 未知的验证标签，返回默认错误信息
		return fmt.Sprintf("%s验证失败(%s)", field, fe.Tag())
	}
}

// translateFieldName 翻译字段名称为中文
func translateFieldName(field string) string {
	fieldNameMap := map[string]string{
		"Username":      "用户名",
		"Email":         "邮箱",
		"Password":      "密码",
		"Phone":         "手机号",
		"OldPassword":   "原密码",
		"NewPassword":   "新密码",
		"Token":         "令牌",
		"RefreshToken":  "刷新令牌",
		"Role":          "角色",
		"Status":        "激活状态",
		"EmailVerified": "邮箱验证状态",
	}

	if translatedName, ok := fieldNameMap[field]; ok {
		return translatedName
	}
	return field
}
