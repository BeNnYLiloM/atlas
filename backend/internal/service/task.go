package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository/postgres"
)

type TaskService struct {
	repo *postgres.TaskRepository
}

func NewTaskService(repo *postgres.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, userID string, input *domain.TaskCreate) (*domain.Task, error) {
	priority := input.Priority
	if priority == "" {
		priority = domain.TaskPriorityMedium
	}

	task := &domain.Task{
		ID:          uuid.New().String(),
		MessageID:   input.MessageID,
		WorkspaceID: input.WorkspaceID,
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

func (s *TaskService) GetByWorkspace(ctx context.Context, workspaceID, status string) ([]*domain.Task, error) {
	return s.repo.GetByWorkspace(ctx, workspaceID, status)
}

func (s *TaskService) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) Update(ctx context.Context, id string, update *domain.TaskUpdate) error {
	return s.repo.Update(ctx, id, update)
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
