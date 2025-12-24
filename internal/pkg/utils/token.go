package utils

import (
	"crypto/rand"
	"encoding/hex"
	math_rand "math/rand/v2"
)

// GenerateRandomToken 生成指定长度的随机 Token
// length: Token 字节长度（实际返回的字符串长度为 length * 2）
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GeneratePasswordResetToken 生成密码重置 Token（32 字节，64 字符）
func GeneratePasswordResetToken() (string, error) {
	return GenerateRandomToken(32)
}

// GenerateEmailVerificationToken 生成邮箱验证 Token（32 字节，64 字符）
func GenerateEmailVerificationToken() (string, error) {
	return GenerateRandomToken(32)
}

// Random 生成一个在 [min, max) 范围内的随机整数
// min: 最小值（包含）
// max: 最大值（不包含）
func Random(min, max int) int {
	return min + math_rand.IntN(max-min)
}
