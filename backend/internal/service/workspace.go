package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

type WorkspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	channelRepo   repository.ChannelRepository
	roleRepo      repository.WorkspaceRoleRepository
	projectRepo   repository.ProjectRepository
}

func NewWorkspaceService(
	workspaceRepo repository.WorkspaceRepository,
	channelRepo repository.ChannelRepository,
	roleRepo repository.WorkspaceRoleRepository,
	projectRepo repository.ProjectRepository,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		channelRepo:   channelRepo,
		roleRepo:      roleRepo,
		projectRepo:   projectRepo,
	}
}

// Create создает новый воркспейс
func (s *WorkspaceService) Create(ctx context.Context, input domain.WorkspaceCreate, ownerID string) (*domain.Workspace, error) {
	workspace := &domain.Workspace{
		ID:      uuid.New().String(),
		Name:    input.Name,
		OwnerID: ownerID,
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Добавляем создателя как владельца
	member := &domain.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		Role:        domain.RoleOwner,
	}
	if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	// Создаем дефолтный канал "general"
	generalChannel := &domain.Channel{
		ID:          uuid.New().String(),
		WorkspaceID: workspace.ID,
		Name:        "general",
		Type:        domain.ChannelTypeText,
		IsPrivate:   false,
	}
	if err := s.channelRepo.Create(ctx, generalChannel); err != nil {
		return nil, err
	}

	// Создаём системные роли (owner, admin, @everyone)
	roleService := NewWorkspaceRoleService(s.roleRepo, s.workspaceRepo)
	if err := roleService.InitSystemRoles(ctx, workspace.ID); err != nil {
		return nil, err
	}

	return workspace, nil
}

// GetByID возвращает воркспейс по ID
func (s *WorkspaceService) GetByID(ctx context.Context, workspaceID, userID string) (*domain.Workspace, error) {
	// Проверяем членство
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	return workspace, nil
}

// GetUserWorkspaces возвращает все воркспейсы пользователя
func (s *WorkspaceService) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return s.workspaceRepo.GetByUserID(ctx, userID)
}

// AddMember добавляет участника в воркспейс
func (s *WorkspaceService) AddMember(ctx context.Context, workspaceID, userID, role, actorID string) error {
	// Проверяем права актора
	actorMember, err := s.workspaceRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	if actorMember == nil {
		return ErrNotMember
	}

	// Только owner и admin могут добавлять участников
	if actorMember.Role != domain.RoleOwner && actorMember.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	member := &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	}
	return s.workspaceRepo.AddMember(ctx, member)
}

// GetMembers возвращает список участников воркспейса
func (s *WorkspaceService) GetMembers(ctx context.Context, workspaceID, userID string) ([]*domain.WorkspaceMember, error) {
	// Проверяем членство
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	return s.workspaceRepo.GetMembers(ctx, workspaceID)
}

// Update обновляет настройки воркспейса
func (s *WorkspaceService) Update(ctx context.Context, workspaceID string, input domain.WorkspaceUpdate, actorID string) (*domain.Workspace, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}
	if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
		return nil, ErrForbidden
	}

	return s.workspaceRepo.Update(ctx, workspaceID, &input)
}

// UpdateMember изменяет роль/никнейм участника
func (s *WorkspaceService) UpdateMember(ctx context.Context, workspaceID, targetUserID string, input domain.WorkspaceMemberUpdate, actorID string) error {
	actor, err := s.workspaceRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	if actor == nil {
		return ErrNotMember
	}

	target, err := s.workspaceRepo.GetMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotMember
	}

	// Владельца нельзя изменить
	if target.Role == domain.RoleOwner {
		return ErrForbidden
	}

	// Только owner и admin могут менять роль
	if input.Role != nil && actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	// Admin не может назначить admin-роль (только owner)
	if input.Role != nil && *input.Role == domain.RoleAdmin && actor.Role != domain.RoleOwner {
		return ErrForbidden
	}

	return s.workspaceRepo.UpdateMember(ctx, workspaceID, targetUserID, &input)
}

// RemoveMember исключает участника из воркспейса
func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID, targetUserID, actorID string) error {
	actor, err := s.workspaceRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	if actor == nil {
		return ErrNotMember
	}

	target, err := s.workspaceRepo.GetMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotMember
	}

	// Владельца нельзя исключить
	if target.Role == domain.RoleOwner {
		return ErrForbidden
	}

	// Admin не может исключить другого Admin
	if actor.Role == domain.RoleAdmin && target.Role == domain.RoleAdmin {
		return ErrForbidden
	}

	if actor.Role != domain.RoleOwner && actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	if err := s.workspaceRepo.RemoveMember(ctx, workspaceID, targetUserID); err != nil {
		return err
	}
	// Каскадно удаляем из всех проектов этого воркспейса
	if err := s.projectRepo.RemoveMemberFromAllProjects(ctx, workspaceID, targetUserID); err != nil {
		return fmt.Errorf("remove from projects: %w", err)
	}
	return nil
}

// Delete удаляет воркспейс
func (s *WorkspaceService) Delete(ctx context.Context, workspaceID, userID string) error {
	// Проверяем права
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	if member == nil || member.Role != domain.RoleOwner {
		return ErrForbidden
	}

	return s.workspaceRepo.Delete(ctx, workspaceID)
}

