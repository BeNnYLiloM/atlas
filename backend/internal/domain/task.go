package domain

import "time"

type TaskStatus string
type TaskPriority string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"

	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

type Task struct {
	ID          string       `json:"id" db:"id"`
	MessageID   *string      `json:"message_id" db:"message_id"`
	WorkspaceID string       `json:"workspace_id" db:"workspace_id"`
	ProjectID   *string      `json:"project_id" db:"project_id"`
	Title       string       `json:"title" db:"title"`
	Description *string      `json:"description" db:"description"`
	Status      TaskStatus   `json:"status" db:"status"`
	Priority    TaskPriority `json:"priority" db:"priority"`
	AssigneeID  *string      `json:"assignee_id" db:"assignee_id"`
	ReporterID  *string      `json:"reporter_id" db:"reporter_id"`
	DueDate     *time.Time   `json:"due_date" db:"due_date"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`

	// Поля для отображения (не в БД)
	Assignee *User `json:"assignee,omitempty" db:"-"`
	Reporter *User `json:"reporter,omitempty" db:"-"`
}

type TaskCreate struct {
	MessageID   *string      `json:"message_id"`
	WorkspaceID string       `json:"workspace_id" binding:"required"`
	ProjectID   *string      `json:"project_id"`
	Title       string       `json:"title" binding:"required,min=1,max=255"`
	Description *string      `json:"description"`
	Priority    TaskPriority `json:"priority"`
	AssigneeID  *string      `json:"assignee_id"`
	DueDate     *time.Time   `json:"due_date"`
}

type TaskUpdate struct {
	Status     *TaskStatus   `json:"status"`
	Priority   *TaskPriority `json:"priority"`
	AssigneeID *string       `json:"assignee_id"`
	DueDate    *time.Time    `json:"due_date"`
	Title      *string       `json:"title"`
}
