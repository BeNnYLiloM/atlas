package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

type ProjectService struct {
	projectRepo       repository.ProjectRepository
	workspaceRepo     repository.WorkspaceRepository
	roleRepo          repository.WorkspaceRoleRepository
	channelRepo       repository.ChannelRepository
	channelPermRepo   repository.ChannelPermissionRepository
	channelMemberRepo repository.ChannelMemberRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	workspaceRepo repository.WorkspaceRepository,
	roleRepo repository.WorkspaceRoleRepository,
	channelRepo repository.ChannelRepository,
	channelPermRepo repository.ChannelPermissionRepository,
	channelMemberRepo repository.ChannelMemberRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo:       projectRepo,
		workspaceRepo:     workspaceRepo,
		roleRepo:          roleRepo,
		channelRepo:       channelRepo,
		channelPermRepo:   channelPermRepo,
		channelMemberRepo: channelMemberRepo,
	}
}

// canManageProject проверяет права на управление проектом:
// ws owner, ws member с ManageWorkspace/ManageMembers, или is_lead в проекте.
func (s *ProjectService) canManageProject(
	ctx context.Context,
	project *domain.Project,
	userID string,
	checkLead bool,
) error {
	member, err := s.workspaceRepo.GetMember(ctx, project.WorkspaceID, userID)
	if err != nil || member == nil {
		return ErrNotMember
	}
	if member.Role == domain.RoleOwner {
		return nil
	}
	perms, err := s.roleRepo.GetEffectivePermissions(ctx, project.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("get effective permissions: %w", err)
	}
	if perms.ManageWorkspace || perms.ManageMembers {
		return nil
	}
	if checkLead {
		pm, err := s.projectRepo.GetMember(ctx, project.ID, userID)
		if err != nil {
			return fmt.Errorf("get project member: %w", err)
		}
		if pm != nil && pm.IsLead {
			return nil
		}
	}
	return ErrForbidden
}

// Create создаёт проект и атомарно добавляет создателя как лида.
func (s *ProjectService) Create(ctx context.Context, workspaceID string, input domain.ProjectCreate, actorID string) (*domain.Project, error) {
	// Проверяем членство и право создавать проекты
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}
	if member.Role != domain.RoleOwner {
		perms, err := s.roleRepo.GetEffectivePermissions(ctx, workspaceID, actorID)
		if err != nil {
			return nil, fmt.Errorf("get effective permissions: %w", err)
		}
		if !perms.CreateProjects {
			return nil, ErrForbidden
		}
	}

	project := &domain.Project{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Name:        input.Name,
		Description: input.Description,
		IconURL:     input.IconURL,
		CreatedAt:   time.Now(),
	}
	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	// Создатель всегда добавляется как лид
	pm := &domain.ProjectMember{
		ProjectID: project.ID,
		UserID:    actorID,
		IsLead:    true,
	}
	if err := s.projectRepo.AddMember(ctx, pm); err != nil {
		return nil, fmt.Errorf("add project creator as lead: %w", err)
	}

	return project, nil
}

// GetByID возвращает проект с проверкой: workspace → membership.
func (s *ProjectService) GetByID(ctx context.Context, projectID, userID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	// Проверяем что userID состоит в этом воркспейсе
	member, err := s.workspaceRepo.GetMember(ctx, project.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrForbidden
	}

	// ws owner всегда видит
	if member.Role == domain.RoleOwner {
		return project, nil
	}

	// ViewAllProjects
	perms, err := s.roleRepo.GetEffectivePermissions(ctx, project.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("get effective permissions: %w", err)
	}
	if perms.ViewAllProjects {
		return project, nil
	}

	// Проверяем членство в проекте
	pm, err := s.projectRepo.GetMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if pm == nil {
		return nil, ErrForbidden
	}

	return project, nil
}

// List возвращает проекты воркспейса, доступные userID.
func (s *ProjectService) List(ctx context.Context, workspaceID, userID string) ([]*domain.Project, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotMember
	}

	all, err := s.projectRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if member.Role == domain.RoleOwner {
		return all, nil
	}

	perms, err := s.roleRepo.GetEffectivePermissions(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("get effective permissions: %w", err)
	}
	if perms.ViewAllProjects {
		return all, nil
	}

	// Фильтруем только проекты, в которых состоит пользователь
	result := make([]*domain.Project, 0)
	for _, p := range all {
		pm, err := s.projectRepo.GetMember(ctx, p.ID, userID)
		if err != nil {
			return nil, err
		}
		if pm != nil {
			result = append(result, p)
		}
	}
	return result, nil
}

// Update обновляет настройки проекта.
func (s *ProjectService) Update(ctx context.Context, projectID string, input domain.ProjectUpdate, actorID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, ErrProjectNotFound
	}
	if err := s.canManageProject(ctx, project, actorID, true); err != nil {
		return nil, err
	}
	return s.projectRepo.Update(ctx, projectID, &input)
}

// Delete удаляет проект. Только ws owner, требует force=true.
func (s *ProjectService) Delete(ctx context.Context, projectID, actorID string, force bool) error {
	if !force {
		return ErrForbidden
	}
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	member, err := s.workspaceRepo.GetMember(ctx, project.WorkspaceID, actorID)
	if err != nil || member == nil || member.Role != domain.RoleOwner {
		return ErrForbidden
	}
	return s.projectRepo.Delete(ctx, projectID)
}

// Archive архивирует проект (read-only режим).
func (s *ProjectService) Archive(ctx context.Context, projectID, actorID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	if err := s.canManageProject(ctx, project, actorID, true); err != nil {
		return err
	}
	return s.projectRepo.SetArchived(ctx, projectID, true)
}

// Unarchive снимает архивацию.
func (s *ProjectService) Unarchive(ctx context.Context, projectID, actorID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	if err := s.canManageProject(ctx, project, actorID, true); err != nil {
		return err
	}
	return s.projectRepo.SetArchived(ctx, projectID, false)
}

// GetMembers возвращает участников проекта.
func (s *ProjectService) GetMembers(ctx context.Context, projectID, actorID string) ([]*domain.ProjectMember, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, ErrProjectNotFound
	}
	if _, err := s.GetByID(ctx, projectID, actorID); err != nil {
		return nil, err
	}
	return s.projectRepo.GetMembers(ctx, projectID)
}

// AddMember добавляет участника воркспейса в проект.
// Безопасность: проверяем что targetUserID состоит в том же воркспейсе.
func (s *ProjectService) AddMember(ctx context.Context, projectID, targetUserID, actorID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	if err := s.canManageProject(ctx, project, actorID, true); err != nil {
		return err
	}

	// Проверяем что целевой пользователь состоит в воркспейсе (cross-workspace защита)
	targetMember, err := s.workspaceRepo.GetMember(ctx, project.WorkspaceID, targetUserID)
	if err != nil {
		return fmt.Errorf("check workspace member: %w", err)
	}
	if targetMember == nil {
		return ErrNotMember
	}

	pm := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    targetUserID,
		IsLead:    false,
	}
	if err := s.projectRepo.AddMember(ctx, pm); err != nil {
		return err
	}

	// Добавляем нового участника в публичные каналы проекта
	// и в приватные каналы проекта, к которым у него есть доступ через роль воркспейса.
	// Синхронно: ошибки синхронизации не должны откатывать добавление участника,
	// но должны быть возвращены для диагностики.
	if err := s.syncMemberToProjectChannels(ctx, project, targetUserID); err != nil {
		return fmt.Errorf("sync member to channels (non-fatal): %w", err)
	}

	return nil
}

// syncMemberToProjectChannels добавляет участника проекта в channel_members.
// Публичные каналы проекта — всегда. Приватные — только если у пользователя есть
// роль воркспейса, которая добавлена в права этого канала.
func (s *ProjectService) syncMemberToProjectChannels(ctx context.Context, project *domain.Project, userID string) error {
	channels, err := s.channelRepo.GetByProjectID(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("get project channels: %w", err)
	}

	// Роли пользователя в воркспейсе
	userRoles, err := s.roleRepo.GetMemberRoles(ctx, project.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("get member roles: %w", err)
	}
	userRoleIDs := make(map[string]bool, len(userRoles))
	for _, r := range userRoles {
		userRoleIDs[r.ID] = true
	}

	for _, ch := range channels {
		if !ch.IsPrivate {
			// Публичный канал проекта — добавляем сразу
			if err := s.channelMemberRepo.UpsertMember(ctx, userID, ch.ID); err != nil {
				return fmt.Errorf("upsert member for channel %s: %w", ch.ID, err)
			}
			continue
		}
		// Приватный — проверяем роли канала
		perms, err := s.channelPermRepo.GetPermissions(ctx, ch.ID)
		if err != nil {
			return fmt.Errorf("get channel permissions %s: %w", ch.ID, err)
		}
		for _, rp := range perms.Roles {
			if userRoleIDs[rp.RoleID] {
				if err := s.channelMemberRepo.UpsertMember(ctx, userID, ch.ID); err != nil {
					return fmt.Errorf("upsert member for private channel %s: %w", ch.ID, err)
				}
				break
			}
		}
	}
	return nil
}

// RemoveMember удаляет участника из проекта.
func (s *ProjectService) RemoveMember(ctx context.Context, projectID, targetUserID, actorID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	if err := s.canManageProject(ctx, project, actorID, true); err != nil {
		return err
	}
	return s.projectRepo.RemoveMember(ctx, projectID, targetUserID)
}

// SetLead назначает лида. Только ws owner или ws member с ManageMembers.
func (s *ProjectService) SetLead(ctx context.Context, projectID, targetUserID, actorID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	// Только ws owner или ManageMembers — лиды не могут назначать других лидов
	if err := s.canManageProject(ctx, project, actorID, false); err != nil {
		return err
	}
	// HIGH-4: target должен быть участником проекта
	pm, err := s.projectRepo.GetMember(ctx, projectID, targetUserID)
	if err != nil {
		return fmt.Errorf("check project member: %w", err)
	}
	if pm == nil {
		return ErrNotMember
	}
	return s.projectRepo.SetLead(ctx, projectID, targetUserID, true)
}

// UnsetLead снимает статус лида. Нельзя снять последнего лида.
func (s *ProjectService) UnsetLead(ctx context.Context, projectID, targetUserID, actorID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		return ErrProjectNotFound
	}
	if err := s.canManageProject(ctx, project, actorID, false); err != nil {
		return err
	}
	// HIGH-4: target должен быть участником проекта
	pm, err := s.projectRepo.GetMember(ctx, projectID, targetUserID)
	if err != nil {
		return fmt.Errorf("check project member: %w", err)
	}
	if pm == nil {
		return ErrNotMember
	}
	count, err := s.projectRepo.GetLeadCount(ctx, projectID)
	if err != nil {
		return fmt.Errorf("get lead count: %w", err)
	}
	if count <= 1 {
		return ErrLastLead
	}
	return s.projectRepo.SetLead(ctx, projectID, targetUserID, false)
}

// GetProjectMembersUserIDs возвращает userID всех участников проекта.
// Используется в handlers для WS BroadcastToUsers.
func (s *ProjectService) GetProjectMembersUserIDs(ctx context.Context, projectID string) ([]string, error) {
	members, err := s.projectRepo.GetMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids, nil
}

// GetProjectMembersAndViewAll возвращает userID участников проекта +
// участников воркспейса с ViewAllProjects (для WS broadcast).
func (s *ProjectService) GetProjectMembersAndViewAll(ctx context.Context, projectID, workspaceID string) ([]string, error) {
	memberIDs, err := s.GetProjectMembersUserIDs(ctx, projectID)
	if err != nil {
		return nil, err
	}
	idSet := make(map[string]struct{}, len(memberIDs))
	for _, id := range memberIDs {
		idSet[id] = struct{}{}
	}

	wsMembers, err := s.workspaceRepo.GetMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("get workspace members: %w", err)
	}
	for _, wm := range wsMembers {
		if _, already := idSet[wm.UserID]; already {
			continue
		}
		if wm.Role == domain.RoleOwner {
			idSet[wm.UserID] = struct{}{}
			continue
		}
		perms, err := s.roleRepo.GetEffectivePermissions(ctx, workspaceID, wm.UserID)
		if err != nil {
			continue
		}
		if perms.ViewAllProjects {
			idSet[wm.UserID] = struct{}{}
		}
	}

	result := make([]string, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}
	return result, nil
}
