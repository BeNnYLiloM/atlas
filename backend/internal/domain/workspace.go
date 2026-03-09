package domain

import (
	"time"
)

type Workspace struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	Description *string   `json:"description" db:"description"`
	IconURL     *string   `json:"icon_url" db:"icon_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type WorkspaceMember struct {
	WorkspaceID string           `json:"workspace_id" db:"workspace_id"`
	UserID      string           `json:"user_id" db:"user_id"`
	Role        string           `json:"role" db:"role"` // owner, admin, member
	DisplayName string           `json:"display_name" db:"display_name"`
	AvatarURL   *string          `json:"avatar_url" db:"avatar_url"`
	Nickname    *string          `json:"nickname" db:"nickname"`
	CustomRoles []*WorkspaceRole `json:"custom_roles"`
}

type WorkspaceCreate struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type WorkspaceUpdate struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	IconURL     *string `json:"icon_url"`
}

type WorkspaceMemberUpdate struct {
	Role     *string `json:"role" validate:"omitempty,oneof=admin member"`
	Nickname *string `json:"nickname" validate:"omitempty,max=100"`
}

const (
	RoleOwner  string = "owner"
	RoleAdmin  string = "admin"
	RoleMember string = "member"
)

