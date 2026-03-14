package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

type taskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByWorkspace(ctx context.Context, workspaceID, projectID, status string) ([]*domain.Task, error)
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, id string, update *domain.TaskUpdate) error
	Delete(ctx context.Context, id string) error
}

type TaskService struct {
	repo          taskRepository
	workspaceRepo repository.WorkspaceRepository
	messageRepo   repository.MessageRepository
	channelRepo   repository.ChannelRepository
	roleRepo      repository.WorkspaceRoleRepository
	permRepo      repository.ChannelPermissionRepository
	projectRepo   repository.ProjectRepository
	dmRepo        repository.DMChannelRepository
}

func NewTaskService(
	repo taskRepository,
	workspaceRepo repository.WorkspaceRepository,
	messageRepo repository.MessageRepository,
	channelRepo repository.ChannelRepository,
	roleRepo repository.WorkspaceRoleRepository,
	permRepo repository.ChannelPermissionRepository,
	projectRepo repository.ProjectRepository,
	dmRepo repository.DMChannelRepository,
) *TaskService {
	return &TaskService{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		roleRepo:      roleRepo,
		permRepo:      permRepo,
		projectRepo:   projectRepo,
		dmRepo:        dmRepo,
	}
}

func (s *TaskService) Create(ctx context.Context, userID string, input *domain.TaskCreate) (*domain.Task, error) {
	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, input.WorkspaceID, userID); err != nil {
		return nil, err
	}

	// CRITICAL-2: project_id должен принадлежать тому же workspace, и пользователь должен быть членом проекта
	if input.ProjectID != nil {
		project, err := s.projectRepo.GetByID(ctx, *input.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("get project: %w", err)
		}
		if project == nil || project.WorkspaceID != input.WorkspaceID {
			return nil, ErrForbidden
		}
		pm, err := s.projectRepo.GetMember(ctx, *input.ProjectID, userID)
		if err != nil {
			return nil, fmt.Errorf("check project member: %w", err)
		}
		if pm == nil {
			return nil, ErrForbidden
		}
	}

	if input.AssigneeID != nil {
		if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, input.WorkspaceID, *input.AssigneeID); err != nil {
			return nil, err
		}
	}
	if input.MessageID != nil {
		message, err := s.messageRepo.GetByID(ctx, *input.MessageID)
		if err != nil {
			return nil, err
		}
		if message == nil {
			return nil, ErrMessageNotFound
		}
		channel, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, message.ChannelID, userID)
		if err != nil {
			return nil, err
		}
		if channel.WorkspaceID != input.WorkspaceID {
			return nil, ErrForbidden
		}
	}

	priority := input.Priority
	if priority == "" {
		priority = domain.TaskPriorityMedium
	}

	task := &domain.Task{
		ID:          uuid.New().String(),
		MessageID:   input.MessageID,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TaskStatusTodo,
		Priority:    priority,
		AssigneeID:  input.AssigneeID,
		ReporterID:  &userID,
		DueDate:     input.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetByWorkspace(ctx context.Context, workspaceID, projectID, status, userID string) ([]*domain.Task, error) {
	member, err := ensureWorkspaceMember(ctx, s.workspaceRepo, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	// CRITICAL-3: если запрашиваются задачи конкретного проекта — проверяем membership
	if projectID != "" {
		// ws owner/admin видят всё
		if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
			pm, pmErr := s.projectRepo.GetMember(ctx, projectID, userID)
			if pmErr != nil {
				return nil, fmt.Errorf("check project member: %w", pmErr)
			}
			if pm == nil {
				return nil, ErrForbidden
			}
		}
		// Проверяем что проект принадлежит этому workspace (защита от cross-workspace)
		project, pErr := s.projectRepo.GetByID(ctx, projectID)
		if pErr != nil {
			return nil, fmt.Errorf("get project: %w", pErr)
		}
		if project == nil || project.WorkspaceID != workspaceID {
			return nil, ErrForbidden
		}
	}

	return s.repo.GetByWorkspace(ctx, workspaceID, projectID, status)
}

func (s *TaskService) GetByID(ctx context.Context, id, userID string) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, task.WorkspaceID, userID); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Update(ctx context.Context, id, userID string, update *domain.TaskUpdate) error {
	task, member, err := s.getAuthorizedTask(ctx, id, userID)
	if err != nil {
		return err
	}
	if task.ReporterID == nil || *task.ReporterID != userID {
		isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
		if !isAssignee && member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
			return ErrForbidden
		}
	}
	return s.repo.Update(ctx, id, update)
}

func (s *TaskService) Delete(ctx context.Context, id, userID string) error {
	task, member, err := s.getAuthorizedTask(ctx, id, userID)
	if err != nil {
		return err
	}
	if task.ReporterID == nil || *task.ReporterID != userID {
		if member.Role != domain.RoleOwner && member.Role != domain.RoleAdmin {
			return ErrForbidden
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *TaskService) getAuthorizedTask(ctx context.Context, id, userID string) (*domain.Task, *domain.WorkspaceMember, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if task == nil {
		return nil, nil, ErrTaskNotFound
	}
	member, err := ensureWorkspaceMember(ctx, s.workspaceRepo, task.WorkspaceID, userID)
	if err != nil {
		return nil, nil, err
	}
	return task, member, nil
}
