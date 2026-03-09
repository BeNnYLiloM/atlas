package domain

import "time"

type ChannelCategory struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Position    int       `json:"position" db:"position"`
	IsPrivate   bool      `json:"is_private" db:"is_private"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ChannelCategoryCreate struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	IsPrivate   bool   `json:"is_private"`
}

type ChannelCategoryUpdate struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	Position  *int    `json:"position"`
	IsPrivate *bool   `json:"is_private"`
}

// CategoryPermissions — список ролей и участников с доступом к приватной категории
type CategoryPermissions struct {
	Roles []ChannelAllowedRole `json:"roles"`
	Users []ChannelAllowedUser `json:"users"`
}

