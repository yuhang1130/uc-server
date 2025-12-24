package model

import (
	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	IDBaseModel
	Username     string     `gorm:"type:varchar(50);uniqueIndex:idx_username,composite:deleted_at;not null" json:"username" binding:"required"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Email        string     `gorm:"type:varchar(100);uniqueIndex:idx_email,composite:deleted_at;not null" json:"email" binding:"required,email"`
	Role         UserRoles  `gorm:"type:varchar(50);not null;default:'user'" json:"role" binding:"required"`
	Status       UserStatus `gorm:"type:tinyint;default:1" json:"status" binding:"required"`
}

// TableName 指定表名
func (u *User) TableName() string {
	return "user"
}

// SetPassword 设置加密后的密码
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
