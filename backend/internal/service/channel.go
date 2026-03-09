package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

var channelNameRe = regexp.MustCompile(`[^\p{L}\p{N}\-]`)

// normalizeChannelName приводит имя канала к нижнему регистру, заменяет пробелы на дефис
func normalizeChannelName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, "-")
	name = channelNameRe.ReplaceAllString(name, "")
	return name
}

type ChannelService struct {
	channelRepo       repository.ChannelRepository
	workspaceRepo     repository.WorkspaceRepository
	channelMemberRepo repository.ChannelMemberRepository
	permRepo          repository.ChannelPermissionRepository
	roleRepo          repository.WorkspaceRoleRepository
}

func NewChannelService(
	channelRepo repository.ChannelRepository,
	workspaceRepo repository.WorkspaceRepository,
	channelMemberRepo repository.ChannelMemberRepository,
	permRepo repository.ChannelPermissionRepository,
	roleRepo repository.WorkspaceRoleRepository,
) *ChannelService {
	return &ChannelService{
		channelRepo:       channelRepo,
		workspaceRepo:     workspaceRepo,
		channelMemberRepo: channelMemberRepo,
		permRepo:          permRepo,
		roleRepo:          roleRepo,
	}
}

// Create создает новый канал
func (s *ChannelService) Create(ctx context.Context, input domain.ChannelCreate, userID string) (*domain.Channel, error) {
	// Проверяем членство и права
	member, err := s.workspaceRepo.GetMember(ctx, input.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	// Только owner и admin могут создавать каналы
	if member.Role == domain.RoleMember {
		return nil, ErrForbidden
	}

	channel := &domain.Channel{
		ID:          uuid.New().String(),
		WorkspaceID: input.WorkspaceID,
		Name:        normalizeChannelName(input.Name),
		Type:        input.Type,
		IsPrivate:   input.IsPrivate,
		CategoryID:  input.CategoryID,
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// GetByID возвращает канал по ID
func (s *ChannelService) GetByID(ctx context.Context, channelID, userID string) (*domain.Channel, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	// Проверяем членство в воркспейсе
	member, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	return channel, nil
}

// GetByWorkspaceID возвращает каналы воркспейса, доступные пользователю
func (s *ChannelService) GetByWorkspaceID(ctx context.Context, workspaceID, userID string) ([]*domain.Channel, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	// owner и admin видят все каналы включая приватные
	if member.Role == domain.RoleOwner || member.Role == domain.RoleAdmin {
		return s.channelRepo.GetByWorkspaceID(ctx, workspaceID)
	}

	// Для остальных — получаем их кастомные роли и фильтруем
	wsRoles, err := s.roleRepo.GetMemberRoles(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]string, 0, len(wsRoles))
	for _, r := range wsRoles {
		roleIDs = append(roleIDs, r.ID)
	}

	return s.channelRepo.GetVisibleByWorkspaceID(ctx, workspaceID, userID, roleIDs)
}

// GetByWorkspaceIDWithUnread возвращает каналы с количеством непрочитанных и уровнем уведомлений
func (s *ChannelService) GetByWorkspaceIDWithUnread(ctx context.Context, workspaceID, userID string) ([]*domain.ChannelWithUnread, error) {
	channels, err := s.GetByWorkspaceID(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.channelMemberRepo.GetUnreadCountsForWorkspace(ctx, userID, workspaceID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.ChannelWithUnread, len(channels))
	for i, ch := range channels {
		s := stats[ch.ID]
		notifLevel := s.NotificationLevel
		if notifLevel == "" {
			notifLevel = domain.NotificationAll
		}
		result[i] = &domain.ChannelWithUnread{
			Channel:           ch,
			UnreadCount:       s.UnreadCount,
			MentionCount:      s.MentionCount,
			NotificationLevel: notifLevel,
		}
	}

	return result, nil
}

// Update обновляет настройки канала
func (s *ChannelService) Update(ctx context.Context, channelID string, input domain.ChannelUpdate, userID string) (*domain.Channel, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	// Проверяем права
	member, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}
	if member.Role == domain.RoleMember {
		return nil, ErrForbidden
	}

	if input.Name != nil {
		normalized := normalizeChannelName(*input.Name)
		input.Name = &normalized
	}

	return s.channelRepo.Update(ctx, channelID, &input)
}

// UpdateNotifications обновляет уровень уведомлений текущего пользователя в канале
func (s *ChannelService) UpdateNotifications(ctx context.Context, channelID, userID, level string) error {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	member, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotMember
	}

	return s.channelMemberRepo.UpdateNotificationLevel(ctx, userID, channelID, level)
}

// Delete удаляет канал
func (s *ChannelService) Delete(ctx context.Context, channelID, userID string) error {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// Проверяем права
	member, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotMember
	}
	if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	return s.channelRepo.Delete(ctx, channelID)
}

// MarkAsRead отмечает канал прочитанным до определенного сообщения
func (s *ChannelService) MarkAsRead(ctx context.Context, channelID, userID string, messageID *string) error {
	// Проверяем доступ к каналу
	channel, err := s.GetByID(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if channel == nil {
		return ErrNotFound
	}

	return s.channelMemberRepo.MarkAsRead(ctx, userID, channelID, messageID)
}

// GetUnreadCount возвращает количество непрочитанных сообщений в канале
func (s *ChannelService) GetUnreadCount(ctx context.Context, channelID, userID string) (int, error) {
	_, err := s.GetByID(ctx, channelID, userID)
	if err != nil {
		return 0, err
	}

	return s.channelMemberRepo.GetUnreadCount(ctx, userID, channelID)
}

// GetChannelMembers возвращает участников канала
func (s *ChannelService) GetChannelMembers(ctx context.Context, channelID, userID string) ([]*domain.ChannelMemberInfo, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	member, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	return s.channelMemberRepo.GetMembers(ctx, channelID)
}

// AddChannelMember добавляет участника в приватный канал
func (s *ChannelService) AddChannelMember(ctx context.Context, channelID, targetUserID, actorID string) error {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	actor, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if actor == nil {
		return ErrNotMember
	}
	if actor.Role == domain.RoleMember {
		return ErrForbidden
	}

	// Проверяем что добавляемый — член воркспейса
	target, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotMember
	}

	return s.channelMemberRepo.UpsertMember(ctx, targetUserID, channelID)
}

// RemoveChannelMember удаляет участника из приватного канала
func (s *ChannelService) RemoveChannelMember(ctx context.Context, channelID, targetUserID, actorID string) error {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	actor, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if actor == nil {
		return ErrNotMember
	}
	if actor.Role == domain.RoleMember {
		return ErrForbidden
	}

	return s.channelMemberRepo.RemoveMember(ctx, targetUserID, channelID)
}

// --- Права доступа к каналу ---

func (s *ChannelService) getChannelWithActor(ctx context.Context, channelID, actorID string) (*domain.Channel, *domain.WorkspaceMember, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, nil, err
	}
	if channel == nil {
		return nil, nil, ErrChannelNotFound
	}
	actor, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, actorID)
	if err != nil {
		return nil, nil, err
	}
	if actor == nil {
		return nil, nil, ErrNotMember
	}
	return channel, actor, nil
}

func (s *ChannelService) GetPermissions(ctx context.Context, channelID, actorID string) (*domain.ChannelPermissions, error) {
	if _, _, err := s.getChannelWithActor(ctx, channelID, actorID); err != nil {
		return nil, err
	}
	return s.permRepo.GetPermissions(ctx, channelID)
}

func (s *ChannelService) AddRolePermission(ctx context.Context, channelID string, input domain.AddRoleInput, actorID string) error {
	_, actor, err := s.getChannelWithActor(ctx, channelID, actorID)
	if err != nil {
		return err
	}
	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	return s.permRepo.AddRole(ctx, channelID, input.RoleID)
}

func (s *ChannelService) RemoveRolePermission(ctx context.Context, channelID, roleID, actorID string) error {
	_, actor, err := s.getChannelWithActor(ctx, channelID, actorID)
	if err != nil {
		return err
	}
	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	return s.permRepo.RemoveRole(ctx, channelID, roleID)
}

func (s *ChannelService) AddUserPermission(ctx context.Context, channelID string, input domain.AddUserInput, actorID string) error {
	_, actor, err := s.getChannelWithActor(ctx, channelID, actorID)
	if err != nil {
		return err
	}
	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	return s.permRepo.AddUser(ctx, channelID, input.UserID)
}

func (s *ChannelService) RemoveUserPermission(ctx context.Context, channelID, targetUserID, actorID string) error {
	_, actor, err := s.getChannelWithActor(ctx, channelID, actorID)
	if err != nil {
		return err
	}
	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	return s.permRepo.RemoveUser(ctx, channelID, targetUserID)
}

// GetAllWorkspaceUserIDs возвращает userID всех участников воркспейса
func (s *ChannelService) GetAllWorkspaceUserIDs(ctx context.Context, workspaceID string) ([]string, error) {
	return s.workspaceRepo.GetMemberUserIDs(ctx, workspaceID)
}

// GetRoleUserIDs возвращает userID всех участников воркспейса с данной ролью
func (s *ChannelService) GetRoleUserIDs(ctx context.Context, roleID string) ([]string, error) {
	return s.roleRepo.GetUserIDsByRole(ctx, roleID)
}

// GetChannelsByRole возвращает все приватные каналы, в которых данная роль имеет доступ
func (s *ChannelService) GetChannelsByRole(ctx context.Context, roleID string) ([]*domain.Channel, error) {
	return s.permRepo.GetChannelsByRole(ctx, roleID)
}

// GetAccessibleUserIDs возвращает список userID участников воркспейса, которым виден канал.
// Для публичных — все участники. Для приватных — owner/admin + те у кого явный доступ или нужная роль.
func (s *ChannelService) GetAccessibleUserIDs(ctx context.Context, channel *domain.Channel) ([]string, error) {
	allIDs, err := s.workspaceRepo.GetMemberUserIDs(ctx, channel.WorkspaceID)
	if err != nil {
		return nil, err
	}

	if !channel.IsPrivate {
		return allIDs, nil
	}

	// Получаем права канала
	perms, err := s.permRepo.GetPermissions(ctx, channel.ID)
	if err != nil {
		return nil, err
	}

	// Составляем set roleIDs разрешённых ролей
	allowedRoleIDs := make(map[string]bool, len(perms.Roles))
	for _, r := range perms.Roles {
		allowedRoleIDs[r.RoleID] = true
	}
	allowedUserIDs := make(map[string]bool, len(perms.Users))
	for _, u := range perms.Users {
		allowedUserIDs[u.UserID] = true
	}

	var result []string
	for _, uid := range allIDs {
		member, err := s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, uid)
		if err != nil || member == nil {
			continue
		}
		// owner и admin всегда видят
		if member.Role == domain.RoleOwner || member.Role == domain.RoleAdmin {
			result = append(result, uid)
			continue
		}
		// явный доступ по userID
		if allowedUserIDs[uid] {
			result = append(result, uid)
			continue
		}
		// проверяем роли пользователя
		if len(allowedRoleIDs) > 0 {
			wsRoles, err := s.roleRepo.GetMemberRoles(ctx, channel.WorkspaceID, uid)
			if err == nil {
				for _, r := range wsRoles {
					if allowedRoleIDs[r.ID] {
						result = append(result, uid)
						break
					}
				}
			}
		}
	}
	return result, nil
}

// CanUserWrite — все участники могут писать в публичные каналы
func (s *ChannelService) CanUserWrite(ctx context.Context, channelID, userID string) (bool, error) {
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil || channel == nil {
		return false, err
	}
	_, err = s.workspaceRepo.GetMember(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return false, err
	}
	return true, nil
}
