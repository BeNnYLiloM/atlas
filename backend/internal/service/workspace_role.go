package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

type WorkspaceRoleService struct {
	roleRepo      repository.WorkspaceRoleRepository
	workspaceRepo repository.WorkspaceRepository
}

func NewWorkspaceRoleService(roleRepo repository.WorkspaceRoleRepository, workspaceRepo repository.WorkspaceRepository) *WorkspaceRoleService {
	return &WorkspaceRoleService{roleRepo: roleRepo, workspaceRepo: workspaceRepo}
}

// InitSystemRoles создаёт системные роли owner, admin, @everyone для нового воркспейса
func (s *WorkspaceRoleService) InitSystemRoles(ctx context.Context, workspaceID string) error {
	systemRoles := []domain.WorkspaceRole{
		{
			ID:          uuid.New().String(),
			WorkspaceID: workspaceID,
			Name:        domain.SystemRoleOwner,
			Color:       "#f1c40f",
			Position:    domain.PosOwner,
			IsSystem:    true,
			Permissions: domain.RolePermissions{
				ManageWorkspace: true, ManageRoles: true, ManageChannels: true,
				ManageMembers: true, ViewAuditLog: true,
				SendMessages: true, AttachFiles: true, MentionEveryone: true,
				ManageMessages: true, ViewChannels: true,
				CreateProjects: true, ViewAllProjects: true,
			},
		},
		{
			ID:          uuid.New().String(),
			WorkspaceID: workspaceID,
			Name:        domain.SystemRoleAdmin,
			Color:       "#e74c3c",
			Position:    domain.PosAdmin,
			IsSystem:    true,
			Permissions: domain.RolePermissions{
				ManageWorkspace: true, ManageRoles: true, ManageChannels: true,
				ManageMembers: true, ViewAuditLog: true,
				SendMessages: true, AttachFiles: true, MentionEveryone: true,
				ManageMessages: true, ViewChannels: true,
				CreateProjects: true, ViewAllProjects: true,
			},
		},
		{
			ID:          uuid.New().String(),
			WorkspaceID: workspaceID,
			Name:        domain.SystemRoleEveryone,
			Color:       "#99aab5",
			Position:    domain.PosEveryone,
			IsSystem:    true,
			Permissions: domain.DefaultEveryonePermissions(),
		},
	}

	for i := range systemRoles {
		if err := s.roleRepo.Create(ctx, &systemRoles[i]); err != nil {
			return fmt.Errorf("create system role %s: %w", systemRoles[i].Name, err)
		}
	}
	return nil
}

// getActorMember получает участника и проверяет что он вообще в воркспейсе
func (s *WorkspaceRoleService) getActorMember(ctx context.Context, workspaceID, actorID string) (*domain.WorkspaceMember, error) {
	actor, err := s.workspaceRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if actor == nil {
		return nil, ErrNotMember
	}
	return actor, nil
}

func isPrivileged(role string) bool {
	return role == domain.RoleOwner || role == domain.RoleAdmin
}

// List возвращает все роли воркспейса
func (s *WorkspaceRoleService) List(ctx context.Context, workspaceID, actorID string) ([]*domain.WorkspaceRole, error) {
	if _, err := s.getActorMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	return s.roleRepo.GetByWorkspaceID(ctx, workspaceID)
}

// Create создаёт кастомную роль
func (s *WorkspaceRoleService) Create(ctx context.Context, workspaceID string, input domain.WorkspaceRoleCreate, actorID string) (*domain.WorkspaceRole, error) {
	actor, err := s.getActorMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if !isPrivileged(actor.Role) {
		return nil, ErrForbidden
	}

	color := input.Color
	if color == "" {
		color = "#99aab5"
	}

	// Позиция: чуть ниже admin, выше @everyone
	role := &domain.WorkspaceRole{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Name:        input.Name,
		Color:       color,
		Position:    1, // кастомные роли между 1–899
		IsSystem:    false,
		Permissions: input.Permissions,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// Update обновляет кастомную роль (системные — нельзя)
func (s *WorkspaceRoleService) Update(ctx context.Context, workspaceID, roleID string, update domain.WorkspaceRoleUpdate, actorID string) (*domain.WorkspaceRole, error) {
	actor, err := s.getActorMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if !isPrivileged(actor.Role) {
		return nil, ErrForbidden
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil || role == nil {
		return nil, ErrNotFound
	}
	if role.WorkspaceID != workspaceID {
		return nil, ErrForbidden
	}
	// Системную роль нельзя переименовать, но можно редактировать permissions @everyone
	if role.IsSystem && role.Name != domain.SystemRoleEveryone {
		return nil, ErrForbidden
	}

	return s.roleRepo.Update(ctx, roleID, &update)
}

// UpdateEveryonePermissions — отдельный метод для @everyone, т.к. он системный
func (s *WorkspaceRoleService) UpdateEveryonePermissions(ctx context.Context, workspaceID string, perms domain.RolePermissions, actorID string) (*domain.WorkspaceRole, error) {
	actor, err := s.getActorMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if !isPrivileged(actor.Role) {
		return nil, ErrForbidden
	}

	roles, err := s.roleRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var everyoneID string
	for _, r := range roles {
		if r.Name == domain.SystemRoleEveryone {
			everyoneID = r.ID
			break
		}
	}
	if everyoneID == "" {
		return nil, ErrNotFound
	}

	return s.roleRepo.Update(ctx, everyoneID, &domain.WorkspaceRoleUpdate{Permissions: &perms})
}

// Delete удаляет кастомную роль
func (s *WorkspaceRoleService) Delete(ctx context.Context, workspaceID, roleID string, actorID string) error {
	actor, err := s.getActorMember(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	if !isPrivileged(actor.Role) {
		return ErrForbidden
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil || role == nil {
		return ErrNotFound
	}
	if role.WorkspaceID != workspaceID {
		return ErrForbidden
	}

	return s.roleRepo.Delete(ctx, roleID)
}

// AssignRole назначает роль участнику
func (s *WorkspaceRoleService) AssignRole(ctx context.Context, workspaceID, targetUserID, roleID, actorID string) error {
	actor, err := s.getActorMember(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	if !isPrivileged(actor.Role) {
		return ErrForbidden
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil || role == nil {
		return ErrNotFound
	}
	if role.WorkspaceID != workspaceID {
		return ErrForbidden
	}
	// Нельзя назначить системные роли owner/admin через этот метод
	if role.IsSystem && (role.Name == domain.SystemRoleOwner || role.Name == domain.SystemRoleAdmin) {
		return ErrForbidden
	}

	return s.roleRepo.AssignRole(ctx, workspaceID, targetUserID, roleID)
}

// RevokeRole снимает роль с участника
func (s *WorkspaceRoleService) RevokeRole(ctx context.Context, workspaceID, targetUserID, roleID, actorID string) error {
	actor, err := s.getActorMember(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	if !isPrivileged(actor.Role) {
		return ErrForbidden
	}

	return s.roleRepo.RevokeRole(ctx, workspaceID, targetUserID, roleID)
}

// GetMemberRoles возвращает роли конкретного участника
func (s *WorkspaceRoleService) GetMemberRoles(ctx context.Context, workspaceID, targetUserID, actorID string) ([]*domain.WorkspaceRole, error) {
	if _, err := s.getActorMember(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	return s.roleRepo.GetMemberRoles(ctx, workspaceID, targetUserID)
}
