package validation

import (
	"errors"
	"regexp"
	"unicode"
)

// PasswordPolicy 密码策略配置
type PasswordPolicy struct {
	MinLength      int  // 最小长度
	MaxLength      int  // 最大长度
	RequireUpper   bool // 需要大写字母
	RequireLower   bool // 需要小写字母
	RequireDigit   bool // 需要数字
	RequireSpecial bool // 需要特殊字符
}

// DefaultPasswordPolicy 默认密码策略
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	MaxLength:      64,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: false, // 可选特殊字符
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	return ValidatePasswordWithPolicy(password, DefaultPasswordPolicy)
}

// ValidatePasswordWithPolicy 使用自定义策略验证密码
func ValidatePasswordWithPolicy(password string, policy PasswordPolicy) error {
	// 检查长度
	if len(password) < policy.MinLength {
		return errors.New("密码长度不能少于 " + string(rune(policy.MinLength+'0')) + " 个字符")
	}
	if len(password) > policy.MaxLength {
		return errors.New("密码长度不能超过 " + string(rune(policy.MaxLength+'0')) + " 个字符")
	}

	// 检查字符类型
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 验证是否满足策略要求
	if policy.RequireUpper && !hasUpper {
		return errors.New("密码必须包含至少一个大写字母")
	}
	if policy.RequireLower && !hasLower {
		return errors.New("密码必须包含至少一个小写字母")
	}
	if policy.RequireDigit && !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}
	if policy.RequireSpecial && !hasSpecial {
		return errors.New("密码必须包含至少一个特殊字符")
	}

	return nil
}

// ContainsCommonPassword 检查是否为常见弱密码
func ContainsCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "iloveyou", "master", "sunshine", "ashley",
		"bailey", "shadow", "superman", "qwertyuiop", "welcome",
	}

	lowerPassword := toLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}
	return false
}

// ContainsSequentialChars 检查是否包含连续字符
func ContainsSequentialChars(password string, maxSequence int) bool {
	if maxSequence <= 1 {
		return false
	}

	// 检查连续数字 (123, 456, 789)
	digitPattern := `\d{` + string(rune(maxSequence+'0')) + `,}`
	if matched, _ := regexp.MatchString(digitPattern, password); matched {
		// 检查是否为递增或递减序列
		for i := 0; i < len(password)-maxSequence+1; i++ {
			isSequential := true
			for j := 0; j < maxSequence-1; j++ {
				if !unicode.IsDigit(rune(password[i+j])) {
					isSequential = false
					break
				}
				diff := int(password[i+j+1]) - int(password[i+j])
				if diff != 1 && diff != -1 {
					isSequential = false
					break
				}
			}
			if isSequential {
				return true
			}
		}
	}

	// 检查连续字母 (abc, xyz)
	for i := 0; i < len(password)-maxSequence+1; i++ {
		isSequential := true
		for j := 0; j < maxSequence-1; j++ {
			if !unicode.IsLetter(rune(password[i+j])) {
				isSequential = false
				break
			}
			diff := int(toLower(string(password[i+j+1]))[0]) - int(toLower(string(password[i+j]))[0])
			if diff != 1 && diff != -1 {
				isSequential = false
				break
			}
		}
		if isSequential {
			return true
		}
	}

	return false
}

// ValidatePasswordWithCommonChecks 综合验证密码（包括常见弱密码检查）
func ValidatePasswordWithCommonChecks(password string) error {
	// 基础强度验证
	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}

	// 检查常见弱密码
	if ContainsCommonPassword(password) {
		return errors.New("密码过于简单，请使用更复杂的密码")
	}

	// 检查连续字符（超过3个）
	if ContainsSequentialChars(password, 3) {
		return errors.New("密码不能包含连续的字符序列")
	}

	return nil
}

// toLower 辅助函数：转换为小写
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, char := range s {
		result[i] = unicode.ToLower(char)
	}
	return string(result)
}
