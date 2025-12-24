package validation

import (
	"errors"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
)

// 测试结构体
type TestStruct struct {
	Username string `validate:"required,min=3,max=32,alphanum"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=64"`
}

func TestTranslateValidationError(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name     string
		input    TestStruct
		wantErr  bool
		contains string
	}{
		{
			name: "required field missing - Username",
			input: TestStruct{
				Email:    "test@example.com",
				Password: "Test1234",
			},
			wantErr:  true,
			contains: "用户名不能为空",
		},
		{
			name: "email format invalid",
			input: TestStruct{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "Test1234",
			},
			wantErr:  true,
			contains: "邮箱格式不正确",
		},
		{
			name: "password too short",
			input: TestStruct{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Test12",
			},
			wantErr:  true,
			contains: "密码长度不能少于8个字符",
		},
		{
			name: "username too short",
			input: TestStruct{
				Username: "ab",
				Email:    "test@example.com",
				Password: "Test1234",
			},
			wantErr:  true,
			contains: "用户名长度不能少于3个字符",
		},
		{
			name: "username with special chars",
			input: TestStruct{
				Username: "test@user",
				Email:    "test@example.com",
				Password: "Test1234",
			},
			wantErr:  true,
			contains: "用户名只能包含字母和数字",
		},
		{
			name: "valid input",
			input: TestStruct{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Test1234",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate.Struct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				translated := TranslateValidationError(err)
				if tt.contains != "" && !stringContains(translated, tt.contains) {
					t.Errorf("TranslateValidationError() = %v, want to contain %v", translated, tt.contains)
				}
				t.Logf("Original error: %v", err)
				t.Logf("Translated error: %v", translated)
			}
		})
	}
}

func TestTranslateValidationError_EOF(t *testing.T) {
	err := io.EOF
	translated := TranslateValidationError(err)
	expected := "请求体不能为空"

	if translated != expected {
		t.Errorf("TranslateValidationError(io.EOF) = %v, want %v", translated, expected)
	}
}

func TestTranslateValidationError_GenericError(t *testing.T) {
	err := errors.New("some generic error")
	translated := TranslateValidationError(err)

	if translated != err.Error() {
		t.Errorf("TranslateValidationError() = %v, want %v", translated, err.Error())
	}
}

// stringContains 辅助函数检查字符串是否包含子串
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
