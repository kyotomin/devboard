package auth

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Hash      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshToken struct {
	UserID    uint   `gorm:""`
	TokenHash string `gorm:"type:text"`
	ExpiresAt time.Time
}
