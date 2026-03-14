package repository

import (
	"context"
	"time"

	"github.com/your-org/atlas/backend/internal/domain"
)

// UserRepository - интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetStatusByID(ctx context.Context, userID string) (string, error)
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, userID, newHash string) error
	UpdateLastSeen(ctx context.Context, userID string, t time.Time) error
	DeleteByID(ctx context.Context, userID string) error
}

// AuthSessionRepository - интерфейс для управления refresh-сессиями.
type AuthSessionRepository interface {
	Create(ctx context.Context, session *domain.AuthSession) error
	GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.AuthSession, error)
	Rotate(ctx context.Context, currentSessionID string, nextSession *domain.AuthSession) error
	RevokeByID(ctx context.Context, sessionID string) error
	// RevokeByIDForUser отзывает сессию только если она принадлежит указанному пользователю.
	// Возвращает false если сессия не найдена или принадлежит другому пользователю.
	RevokeByIDForUser(ctx context.Context, sessionID, userID string) (bool, error)
	RevokeFamily(ctx context.Context, familyID string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
	ListActiveByUserID(ctx context.Context, userID string) ([]*domain.AuthSession, error)
}

// WorkspaceRepository - интерфейс для работы с воркспейсами
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *domain.Workspace) error
	GetByID(ctx context.Context, id string) (*domain.Workspace, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Workspace, error)
	Update(ctx context.Context, id string, update *domain.WorkspaceUpdate) (*domain.Workspace, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *domain.WorkspaceMember) error
	GetMembers(ctx context.Context, workspaceID string) ([]*domain.WorkspaceMember, error)
	GetMemberUserIDs(ctx context.Context, workspaceID string) ([]string, error)
	GetMember(ctx context.Context, workspaceID, userID string) (*domain.WorkspaceMember, error)
	UpdateMember(ctx context.Context, workspaceID, userID string, update *domain.WorkspaceMemberUpdate) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
}

// WorkspaceRoleRepository - интерфейс для работы с ролями воркспейса
type WorkspaceRoleRepository interface {
	// CRUD ролей
	Create(ctx context.Context, role *domain.WorkspaceRole) error
	GetByID(ctx context.Context, id string) (*domain.WorkspaceRole, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.WorkspaceRole, error)
	Update(ctx context.Context, id string, update *domain.WorkspaceRoleUpdate) (*domain.WorkspaceRole, error)
	Delete(ctx context.Context, id string) error

	// Назначение ролей участникам
	AssignRole(ctx context.Context, workspaceID, userID, roleID string) error
	RevokeRole(ctx context.Context, workspaceID, userID, roleID string) error
	GetMemberRoles(ctx context.Context, workspaceID, userID string) ([]*domain.WorkspaceRole, error)
	GetUserIDsByRole(ctx context.Context, roleID string) ([]string, error)

	// Эффективные права участника (merge @everyone + все назначенные роли)
	GetEffectivePermissions(ctx context.Context, workspaceID, userID string) (*domain.RolePermissions, error)
}

// ChannelRepository - интерфейс для работы с каналами
type ChannelRepository interface {
	Create(ctx context.Context, channel *domain.Channel) error
	GetByID(ctx context.Context, id string) (*domain.Channel, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.Channel, error)
	// GetVisibleByWorkspaceID — только каналы доступные пользователю:
	// публичные + приватные где у него есть явный доступ или его роль добавлена
	GetVisibleByWorkspaceID(ctx context.Context, workspaceID, userID string, roleIDs []string) ([]*domain.Channel, error)
	// GetByProjectID — все каналы проекта (для внутренних нужд, без access check)
	GetByProjectID(ctx context.Context, projectID string) ([]*domain.Channel, error)
	// GetVisibleByProjectID — только видимые пользователю каналы проекта
	GetVisibleByProjectID(ctx context.Context, projectID, userID string, roleIDs []string) ([]*domain.Channel, error)
	Update(ctx context.Context, id string, update *domain.ChannelUpdate) (*domain.Channel, error)
	Delete(ctx context.Context, id string) error
}

// ChannelPermissionRepository - интерфейс для прав доступа к каналу
type ChannelPermissionRepository interface {
	GetPermissions(ctx context.Context, channelID string) (*domain.ChannelPermissions, error)
	AddRole(ctx context.Context, channelID, roleID string) error
	RemoveRole(ctx context.Context, channelID, roleID string) error
	AddUser(ctx context.Context, channelID, userID string) error
	RemoveUser(ctx context.Context, channelID, userID string) error
	// HasAccess — есть ли у пользователя доступ к приватному каналу
	HasAccess(ctx context.Context, channelID, userID string, wsRoleIDs []string) (bool, error)
	// GetChannelsByRole — каналы где роль добавлена в channel_role_permissions
	GetChannelsByRole(ctx context.Context, roleID string) ([]*domain.Channel, error)
}

// ChannelCategoryRepository - интерфейс для работы с категориями каналов
type ChannelCategoryRepository interface {
	Create(ctx context.Context, category *domain.ChannelCategory) error
	GetByID(ctx context.Context, id string) (*domain.ChannelCategory, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.ChannelCategory, error)
	Update(ctx context.Context, id string, update *domain.ChannelCategoryUpdate) (*domain.ChannelCategory, error)
	Delete(ctx context.Context, id string) error
}

// CategoryPermissionRepository - права доступа к приватным категориям
type CategoryPermissionRepository interface {
	GetPermissions(ctx context.Context, categoryID string) (*domain.CategoryPermissions, error)
	AddRole(ctx context.Context, categoryID, roleID string) error
	RemoveRole(ctx context.Context, categoryID, roleID string) error
	AddUser(ctx context.Context, categoryID, userID string) error
	RemoveUser(ctx context.Context, categoryID, userID string) error
	// HasAccess — есть ли у пользователя доступ к приватной категории
	HasAccess(ctx context.Context, categoryID, userID string, wsRoleIDs []string) (bool, error)
	// GetVisibleCategoryIDs — ID категорий доступных пользователю
	GetVisibleCategoryIDs(ctx context.Context, workspaceID, userID string, wsRoleIDs []string) ([]string, error)
}

// SearchResult — результат поиска, независимый от реализации (Postgres, ES и т.д.)
type SearchResult struct {
	Message   *domain.Message `json:"message"`
	Rank      float32         `json:"rank"`
	Highlight string          `json:"highlight"`
}

// SearchFilter — параметры поиска.
type SearchFilter struct {
	Query       string
	WorkspaceID string
	ChannelID   string
	UserID      string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

// SearchRepository — абстракция поискового бэкенда.
// Текущая реализация: PostgreSQL FTS + trigram.
// Будущая: Elasticsearch, Typesense и др. — без изменений в сервисе.
type SearchRepository interface {
	Search(ctx context.Context, filter SearchFilter) ([]*SearchResult, int, error)
}

// ProjectRepository - интерфейс для работы с проектами
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.Project, error)
	Update(ctx context.Context, id string, update *domain.ProjectUpdate) (*domain.Project, error)
	Delete(ctx context.Context, id string) error
	SetArchived(ctx context.Context, id string, archived bool) error

	AddMember(ctx context.Context, member *domain.ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID string) error
	GetMembers(ctx context.Context, projectID string) ([]*domain.ProjectMember, error)
	GetMember(ctx context.Context, projectID, userID string) (*domain.ProjectMember, error)
	RemoveMemberFromAllProjects(ctx context.Context, workspaceID, userID string) error

	SetLead(ctx context.Context, projectID, userID string, isLead bool) error
	GetLeadCount(ctx context.Context, projectID string) (int, error)
}

// DMChannelRepository - интерфейс для работы с личными сообщениями
type DMChannelRepository interface {
	// GetOrCreate атомарно возвращает существующий DM-канал или создаёт новый.
	// Транзакция включает: channels + dm_channels + channel_members (оба участника).
	GetOrCreate(ctx context.Context, workspaceID, userID1, userID2 string) (*domain.Channel, error)
	// GetByUserID возвращает все DM-диалоги пользователя с данными собеседника.
	GetByUserID(ctx context.Context, workspaceID, userID string) ([]*domain.DMChannel, error)
	// IsMember проверяет что пользователь является участником DM-канала.
	// Используется в CanAccessChannel для блокировки IDOR.
	IsMember(ctx context.Context, channelID, userID string) (bool, error)
}

// MessageRepository - интерфейс для работы с сообщениями
type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) error
	GetByID(ctx context.Context, id string) (*domain.Message, error)
	GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error)
	GetThreadMessages(ctx context.Context, parentID string) ([]*domain.Message, error)
	Update(ctx context.Context, message *domain.Message) error
	UpdateCallStatus(ctx context.Context, messageID, callerID, status string, durationSec *int, allowedFromStatuses []string) error
	Delete(ctx context.Context, id string) error
}

// ChannelMemberRepository - интерфейс для работы с участниками каналов
type ChannelMemberRepository interface {
	UpsertMember(ctx context.Context, userID, channelID string) error
	RemoveMember(ctx context.Context, userID, channelID string) error
	GetMembers(ctx context.Context, channelID string) ([]*domain.ChannelMemberInfo, error)
	MarkAsRead(ctx context.Context, userID, channelID string, messageID *string) error
	GetUnreadCount(ctx context.Context, userID, channelID string) (int, error)
	GetUnreadCountsForWorkspace(ctx context.Context, userID, workspaceID string) (map[string]domain.ChannelStats, error)
	GetLastReadMessageID(ctx context.Context, userID, channelID string) (*string, error)
	UpdateNotificationLevel(ctx context.Context, userID, channelID, level string) error
	GetNotificationLevel(ctx context.Context, userID, channelID string) (string, error)
	GetLastMessageAt(ctx context.Context, userID, channelID string) (*time.Time, error)
	SetLastMessageAt(ctx context.Context, userID, channelID string) error

	// Thread read state
	MarkThreadAsRead(ctx context.Context, userID, parentMessageID string, lastMessageID *string) error
	GetThreadUnreadCount(ctx context.Context, userID, parentMessageID string) (int, error)
}
