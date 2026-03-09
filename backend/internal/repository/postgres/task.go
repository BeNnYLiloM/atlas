package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, message_id, workspace_id, title, description, status, priority, assignee_id, reporter_id, due_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`
	_, err := r.db.Exec(ctx, query,
		task.ID, task.MessageID, task.WorkspaceID, task.Title, task.Description,
		task.Status, task.Priority, task.AssigneeID, task.ReporterID,
		task.DueDate, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("TaskRepository.Create: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT t.id, t.message_id, t.workspace_id, t.title, t.description, t.status, t.priority,
			   t.assignee_id, t.reporter_id, t.due_date, t.created_at, t.updated_at
		FROM tasks t WHERE t.id = $1
	`
	task := &domain.Task{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.MessageID, &task.WorkspaceID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.AssigneeID, &task.ReporterID,
		&task.DueDate, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("TaskRepository.GetByID: %w", err)
	}
	return task, nil
}

func (r *TaskRepository) GetByWorkspace(ctx context.Context, workspaceID string, status string) ([]*domain.Task, error) {
	where := "t.workspace_id = $1"
	args := []interface{}{workspaceID}

	if status != "" {
		where += " AND t.status = $2"
		args = append(args, status)
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.message_id, t.workspace_id, t.title, t.description, t.status, t.priority,
			   t.assignee_id, t.reporter_id, t.due_date, t.created_at, t.updated_at
		FROM tasks t WHERE %s
		ORDER BY 
			CASE t.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END,
			t.created_at DESC
	`, where)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("TaskRepository.GetByWorkspace: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(
			&task.ID, &task.MessageID, &task.WorkspaceID, &task.Title, &task.Description,
			&task.Status, &task.Priority, &task.AssigneeID, &task.ReporterID,
			&task.DueDate, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("TaskRepository.GetByWorkspace scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, id string, update *domain.TaskUpdate) error {
	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if update.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *update.Status)
		argIdx++
	}
	if update.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *update.Priority)
		argIdx++
	}
	if update.AssigneeID != nil {
		sets = append(sets, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, *update.AssigneeID)
		argIdx++
	}
	if update.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *update.Title)
		argIdx++
	}
	if update.DueDate != nil {
		sets = append(sets, fmt.Sprintf("due_date = $%d", argIdx))
		args = append(args, *update.DueDate)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(sets, ", "), argIdx)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("TaskRepository.Update: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}
