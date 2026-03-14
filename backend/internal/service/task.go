package service

import (
	"context"
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
}

func NewTaskService(
	repo taskRepository,
	workspaceRepo repository.WorkspaceRepository,
	messageRepo repository.MessageRepository,
	channelRepo repository.ChannelRepository,
	roleRepo repository.WorkspaceRoleRepository,
	permRepo repository.ChannelPermissionRepository,
	projectRepo repository.ProjectRepository,
) *TaskService {
	return &TaskService{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		roleRepo:      roleRepo,
		permRepo:      permRepo,
		projectRepo:   projectRepo,
	}
}

func (s *TaskService) Create(ctx context.Context, userID string, input *domain.TaskCreate) (*domain.Task, error) {
	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, input.WorkspaceID, userID); err != nil {
		return nil, err
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
		channel, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, message.ChannelID, userID)
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
	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, workspaceID, userID); err != nil {
		return nil, err
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
