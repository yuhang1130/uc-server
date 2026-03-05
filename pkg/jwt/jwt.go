package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yuhang1130/gin-server/internal/model"
)

const (
	__defaultExpire = 24 * time.Hour
	__deaultIssuer  = "gin-app"
	__deaultSecret  = "sb3MRHerqNjqjOp0xgAPQPCl6HF3dy5z"
)

// JWT -.
type JWT interface {
	GenerateAccessToken(userID uint64, tenantID uint64, role model.UserRoles) (string, string, int64, error)
	GenerateRefreshToken(userID uint64) (string, error)
	ParseAccessToken(tokenString string) (*Claims, error)
	ParseRefreshToken(tokenString string) (jwt.MapClaims, error)
}

// jwtImpl -.
type jwtImpl struct {
	expire time.Duration
	issuer string
	secret string
}

// NewJWT -.
func NewJWT(opts ...Option) JWT {
	j := &jwtImpl{
		expire: __defaultExpire,
		issuer: __deaultIssuer,
		secret: __deaultSecret,
	}

	// Custom options
	for _, opt := range opts {
		opt(j)
	}

	return j
}

// Claims -.
type Claims struct {
	UserID   uint64          `json:"user_id"`
	TenantID uint64          `json:"tenant_id"`
	Role     model.UserRoles `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成访问令牌
func (j *jwtImpl) GenerateAccessToken(userID uint64, tenantID uint64, role model.UserRoles) (string, string, int64, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(j.secret))
	return claims.ID, accessToken, claims.ExpiresAt.UnixMilli(), err
}

// GenerateRefreshToken 生成刷新令牌
func (j *jwtImpl) GenerateRefreshToken(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7天有效期
		"iat":     time.Now().Unix(),
		"iss":     j.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// ParseAccessToken 解析访问令牌
func (j *jwtImpl) ParseAccessToken(t string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(t, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// ParseRefreshToken 解析刷新令牌
func (j *jwtImpl) ParseRefreshToken(t string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(t, func(t *jwt.Token) (any, error) {
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
