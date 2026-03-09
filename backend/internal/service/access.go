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
	channelID, userID string,
) (*domain.Channel, *domain.WorkspaceMember, error) {
	channel, err := channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, nil, err
	}
	if channel == nil {
		return nil, nil, ErrChannelNotFound
	}

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
