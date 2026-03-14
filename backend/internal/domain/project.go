package domain

import "time"

type Project struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	IconURL     *string   `json:"icon_url" db:"icon_url"`
	IsArchived  bool      `json:"is_archived" db:"is_archived"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ProjectMember struct {
	ProjectID string    `json:"project_id" db:"project_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	IsLead    bool      `json:"is_lead" db:"is_lead"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`

	// Поля из JOIN с users (опционально)
	DisplayName string  `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

type ProjectCreate struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IconURL     *string `json:"icon_url"`
}

type ProjectUpdate struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description   *string `json:"description" binding:"omitempty,max=500"`
	IconURL       *string `json:"icon_url"`
	ClearIconURL  bool    `json:"clear_icon_url"`
}

type ProjectMemberAdd struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}
