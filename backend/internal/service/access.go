package service

import (
	"context"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

func ensureWorkspaceMember(
	ctx context.Context,
	workspaceRepo repository.WorkspaceRepository,
	workspaceID, userID string,
) (*domain.WorkspaceMember, error) {
	member, err := workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}
	return member, nil
}

func getAccessibleChannel(
	ctx context.Context,
	channelRepo repository.ChannelRepository,
	workspaceRepo repository.WorkspaceRepository,
	roleRepo repository.WorkspaceRoleRepository,
	permRepo repository.ChannelPermissionRepository,
	projectRepo repository.ProjectRepository,
	dmRepo repository.DMChannelRepository,
	channelID, userID string,
) (*domain.Channel, *domain.WorkspaceMember, error) {
	channel, err := channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, nil, err
	}
	if channel == nil {
		return nil, nil, ErrChannelNotFound
	}

	// DM-канал: доступ только через channel_members, без fallback на workspace-роли.
	// Даже workspace owner не может читать чужие DM.
	if channel.Type == domain.ChannelTypeDM {
		isMember, err := dmRepo.IsMember(ctx, channelID, userID)
		if err != nil {
			return nil, nil, err
		}
		if !isMember {
			return nil, nil, ErrForbidden
		}
		return channel, nil, nil
	}

	// Канал принадлежит проекту — отдельная ветка проверки доступа
	if channel.ProjectID != nil {
		return getAccessibleProjectChannel(ctx, channel, workspaceRepo, roleRepo, projectRepo, userID)
	}

	// Обычный воркспейс-канал — существующая логика
	member, err := ensureWorkspaceMember(ctx, workspaceRepo, channel.WorkspaceID, userID)
	if err != nil {
		return nil, nil, err
	}

	if !channel.IsPrivate || member.Role == domain.RoleOwner || member.Role == domain.RoleAdmin {
		return channel, member, nil
	}

	wsRoles, err := roleRepo.GetMemberRoles(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return nil, nil, err
	}

	roleIDs := make([]string, 0, len(wsRoles))
	for _, role := range wsRoles {
		roleIDs = append(roleIDs, role.ID)
	}

	hasAccess, err := permRepo.HasAccess(ctx, channel.ID, userID, roleIDs)
	if err != nil {
		return nil, nil, err
	}
	if !hasAccess {
		return nil, nil, ErrForbidden
	}

	return channel, member, nil
}

// getAccessibleProjectChannel проверяет доступ к каналу проекта.
// Порядок: ws owner → ViewAllProjects → project_members.
func getAccessibleProjectChannel(
	ctx context.Context,
	channel *domain.Channel,
	workspaceRepo repository.WorkspaceRepository,
	roleRepo repository.WorkspaceRoleRepository,
	projectRepo repository.ProjectRepository,
	userID string,
) (*domain.Channel, *domain.WorkspaceMember, error) {
	member, err := ensureWorkspaceMember(ctx, workspaceRepo, channel.WorkspaceID, userID)
	if err != nil {
		return nil, nil, err
	}

	// ws owner всегда имеет доступ
	if member.Role == domain.RoleOwner {
		return channel, member, nil
	}

	// Проверяем ViewAllProjects в эффективных правах
	perms, err := roleRepo.GetEffectivePermissions(ctx, channel.WorkspaceID, userID)
	if err != nil {
		return nil, nil, err
	}
	if perms.ViewAllProjects {
		return channel, member, nil
	}

	// Проверяем членство в проекте
	pm, err := projectRepo.GetMember(ctx, *channel.ProjectID, userID)
	if err != nil {
		return nil, nil, err
	}
	if pm == nil {
		return nil, nil, ErrForbidden
	}

	return channel, member, nil
}
