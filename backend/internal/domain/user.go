package domain

import (
	"time"
)

type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusAway    UserStatus = "away"
	UserStatusDND     UserStatus = "dnd"
	UserStatusOffline UserStatus = "offline"
)

type User struct {
	ID           string     `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	DisplayName  string     `json:"display_name" db:"display_name"`
	AvatarURL    *string    `json:"avatar_url" db:"avatar_url"`
	Status       UserStatus `json:"status" db:"status"`
	CustomStatus *string    `json:"custom_status" db:"custom_status"`
	LastSeen     *time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type UserCreate struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserUpdate struct {
	DisplayName  *string     `json:"display_name" validate:"omitempty,min=2,max=100"`
	AvatarURL    *string     `json:"avatar_url"`
	Status       *UserStatus `json:"status"`
	CustomStatus *string     `json:"custom_status"`
}

type UserChangePassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
