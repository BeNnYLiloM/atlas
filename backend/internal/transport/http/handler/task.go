package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input domain.TaskCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, &input)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": task})
}

// List GET /api/v1/tasks?workspace_id=uuid&status=todo
func (h *TaskHandler) List(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		response.BadRequest(c, "workspace_id is required")
		return
	}

	tasks, err := h.taskService.GetByWorkspace(c.Request.Context(), workspaceID, c.Query("status"), middleware.GetUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}

	if tasks == nil {
		tasks = []*domain.Task{}
	}
	response.Success(c, tasks)
}

// Update PATCH /api/v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var update domain.TaskUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.taskService.Update(c.Request.Context(), id, middleware.GetUserID(c), &update); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"updated": true})
}

// Delete DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.taskService.Delete(c.Request.Context(), id, middleware.GetUserID(c)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
