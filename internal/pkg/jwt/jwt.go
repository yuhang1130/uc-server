package jwt

import (
	"errors"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yuhang1130/gin-server/internal/model"
)

// JWTUtil JWT工具
type JWTUtil struct {
	secretKey string
	expiresIn int
}

// Claims 自定义声明
type Claims struct {
	UserID   uint64          `json:"user_id"`
	TenantID uint64          `json:"tenant_id"`
	Role     model.UserRoles `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTUtil 创建JWT工具实例
func NewJWTUtil(secretKey string, expiresIn int) *JWTUtil {
	if expiresIn == 0 {
		expiresIn = 24 * 3600 // 默认24小时
	}
	return &JWTUtil{
		secretKey: secretKey,
		expiresIn: expiresIn,
	}
}

// GenerateAccessToken 生成访问令牌
func (j *JWTUtil) GenerateAccessToken(userID uint64, tenantID uint64, role model.UserRoles) (string, string, time.Time, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(j.secretKey))
	return claims.ID, accessToken, claims.ExpiresAt.Time, err
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTUtil) GenerateRefreshToken(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7天有效期
		"iat":     time.Now().Unix(),
		"iss":     "user-service",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ParseAccessToken 解析访问令牌
func (j *JWTUtil) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ParseRefreshToken 解析刷新令牌
func (j *JWTUtil) ParseRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateTokenID 生成token ID
func GenerateTokenID() string {
	return "token-" + time.Now().Format("20060102150405") + "-" + string(rune(rand.Intn(90)+65))
}

// GenerateTokenIDFromToken 从token生成token ID
func GenerateTokenIDFromToken(tokenString string) string {
	return "token-" + tokenString[:10]
}
