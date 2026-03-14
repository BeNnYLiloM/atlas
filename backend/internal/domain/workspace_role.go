package domain

import "time"

// RolePermissions — набор прав для роли
type RolePermissions struct {
	// Управление воркспейсом
	ManageWorkspace bool `json:"manage_workspace"` // редактировать имя/описание воркспейса
	ManageRoles     bool `json:"manage_roles"`     // создавать/редактировать роли ниже своей позиции
	ManageChannels  bool `json:"manage_channels"`  // создавать/редактировать/удалять каналы
	ManageMembers   bool `json:"manage_members"`   // кикать участников, назначать роли ниже себя
	ViewAuditLog    bool `json:"view_audit_log"`   // просматривать журнал действий

	// Каналы
	SendMessages    bool `json:"send_messages"`    // отправлять сообщения
	AttachFiles     bool `json:"attach_files"`     // прикреплять файлы
	MentionEveryone bool `json:"mention_everyone"` // упоминать @everyone
	ManageMessages  bool `json:"manage_messages"`  // удалять/редактировать чужие сообщения

	// Видимость
	ViewChannels bool `json:"view_channels"` // видеть публичные каналы

	// Проекты
	CreateProjects  bool `json:"create_projects"`   // создавать проекты в воркспейсе
	ViewAllProjects bool `json:"view_all_projects"`  // видеть все проекты воркспейса
}

// DefaultEveryonePermissions — права @everyone по умолчанию
func DefaultEveryonePermissions() RolePermissions {
	return RolePermissions{
		ViewChannels: true,
		SendMessages: true,
		AttachFiles:  true,
	}
}

// WorkspaceRole — роль воркспейса
type WorkspaceRole struct {
	ID          string          `json:"id" db:"id"`
	WorkspaceID string          `json:"workspace_id" db:"workspace_id"`
	Name        string          `json:"name" db:"name"`
	Color       string          `json:"color" db:"color"`
	Position    int             `json:"position" db:"position"`
	IsSystem    bool            `json:"is_system" db:"is_system"`
	Permissions RolePermissions `json:"permissions" db:"permissions"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`

	// Количество участников с этой ролью (из JOIN, опционально)
	MemberCount int `json:"member_count,omitempty" db:"member_count"`
}

// WorkspaceRoleCreate — входные данные для создания роли
type WorkspaceRoleCreate struct {
	Name        string          `json:"name" binding:"required,min=1,max=100"`
	Color       string          `json:"color" binding:"omitempty,max=7"`
	Permissions RolePermissions `json:"permissions"`
}

// WorkspaceRoleUpdate — входные данные для обновления роли
type WorkspaceRoleUpdate struct {
	Name        *string          `json:"name" binding:"omitempty,min=1,max=100"`
	Color       *string          `json:"color" binding:"omitempty,max=7"`
	Permissions *RolePermissions `json:"permissions"`
}

// Системные роли (имена)
const (
	SystemRoleOwner    = "owner"
	SystemRoleAdmin    = "admin"
	SystemRoleEveryone = "@everyone"
)

// Системные позиции
const (
	PosOwner    = 1000
	PosAdmin    = 900
	PosEveryone = 0 // всегда внизу
)
