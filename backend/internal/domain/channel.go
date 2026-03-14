package domain

import (
	"time"
)

type Channel struct {
	ID              string    `json:"id" db:"id"`
	WorkspaceID     string    `json:"workspace_id" db:"workspace_id"`
	Name            string    `json:"name" db:"name"`
	Type            string    `json:"type" db:"type"` // text, voice
	IsPrivate       bool      `json:"is_private" db:"is_private"`
	Topic           *string   `json:"topic" db:"topic"`
	SlowmodeSeconds int       `json:"slowmode_seconds" db:"slowmode_seconds"`
	Position        int       `json:"position" db:"position"`
	CategoryID      *string   `json:"category_id" db:"category_id"`
	ProjectID      *string   `json:"project_id" db:"project_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type ChannelCreate struct {
	WorkspaceID string  `json:"workspace_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Type        string  `json:"type" validate:"required,oneof=text voice"`
	IsPrivate   bool    `json:"is_private"`
	CategoryID  *string `json:"category_id"`
	ProjectID   *string `json:"project_id"`
}

type ChannelUpdate struct {
	Name            *string `json:"name" validate:"omitempty,min=2,max=100"`
	Topic           *string `json:"topic" validate:"omitempty,max=1024"`
	IsPrivate       *bool   `json:"is_private"`
	SlowmodeSeconds *int    `json:"slowmode_seconds" validate:"omitempty,min=0,max=21600"`
	CategoryID      *string `json:"category_id"`
}

const (
	ChannelTypeText  = "text"
	ChannelTypeVoice = "voice"
)

// ChannelAllowedRole — роль воркспейса с доступом к каналу
type ChannelAllowedRole struct {
	ChannelID string `json:"channel_id"`
	RoleID    string `json:"role_id"`
	RoleName  string `json:"role_name"`
	RoleColor string `json:"role_color"`
}

// ChannelAllowedUser — участник с индивидуальным доступом к каналу
type ChannelAllowedUser struct {
	ChannelID   string  `json:"channel_id"`
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// ChannelPermissions — список ролей и участников с доступом к каналу
type ChannelPermissions struct {
	Roles []ChannelAllowedRole `json:"roles"`
	Users []ChannelAllowedUser `json:"users"`
}

// AddRoleInput — добавить роль в доступ к каналу
type AddRoleInput struct {
	RoleID string `json:"role_id" binding:"required"`
}

// AddUserInput — добавить участника в доступ к каналу
type AddUserInput struct {
	UserID string `json:"user_id" binding:"required"`
}

