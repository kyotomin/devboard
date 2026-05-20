package auth

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null`
	Email     string `gorm:"unique;not null`
	Hash      string `gorm:"not null`
	CreatedAt time.Time
	UpdatedAt time.Time
}
