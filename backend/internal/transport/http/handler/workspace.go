package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
	fileService      *service.FileService
	wsHub            *ws.Hub
}

func NewWorkspaceHandler(workspaceService *service.WorkspaceService, fileService *service.FileService, wsHub *ws.Hub) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		fileService:      fileService,
		wsHub:            wsHub,
	}
}

// Create создает новый воркспейс
func (h *WorkspaceHandler) Create(c *gin.Context) {
	var input domain.WorkspaceCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	workspace, err := h.workspaceService.Create(c.Request.Context(), input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, workspace)
}

// GetByID возвращает воркспейс по ID
func (h *WorkspaceHandler) GetByID(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	workspace, err := h.workspaceService.GetByID(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, workspace)
}

// GetUserWorkspaces возвращает все воркспейсы пользователя
func (h *WorkspaceHandler) GetUserWorkspaces(c *gin.Context) {
	userID := middleware.GetUserID(c)

	workspaces, err := h.workspaceService.GetUserWorkspaces(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, workspaces)
}

// GetMembers возвращает участников воркспейса
func (h *WorkspaceHandler) GetMembers(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	members, err := h.workspaceService.GetMembers(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, members)
}

// Update обновляет настройки воркспейса
func (h *WorkspaceHandler) Update(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.WorkspaceUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	workspace, err := h.workspaceService.Update(c.Request.Context(), workspaceID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToWorkspace(workspaceID, "workspace_updated", workspace, "")

	response.Success(c, workspace)
}

type AddMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member"`
}

// AddMember добавляет участника в воркспейс
func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.workspaceService.AddMember(c.Request.Context(), workspaceID, input.UserID, input.Role, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Broadcast добавления участника в workspace
	h.wsHub.BroadcastToWorkspace(
		workspaceID,
		"member_added",
		map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      input.UserID,
			"role":         input.Role,
		},
		"", // Здесь не исключаем, все должны знать о новом участнике
	)

	response.NoContent(c)
}

// UpdateMember изменяет роль или никнейм участника
func (h *WorkspaceHandler) UpdateMember(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	var input domain.WorkspaceMemberUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.workspaceService.UpdateMember(c.Request.Context(), workspaceID, targetUserID, input, actorID); err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToWorkspace(workspaceID, "member_updated", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      targetUserID,
		"role":         input.Role,
		"nickname":     input.Nickname,
	}, "")

	response.NoContent(c)
}

// RemoveMember исключает участника из воркспейса
func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	if err := h.workspaceService.RemoveMember(c.Request.Context(), workspaceID, targetUserID, actorID); err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToWorkspace(workspaceID, "member_removed", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      targetUserID,
	}, "")

	response.NoContent(c)
}

// Delete удаляет воркспейс
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	err := h.workspaceService.Delete(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// UploadIcon загружает иконку воркспейса
func (h *WorkspaceHandler) UploadIcon(c *gin.Context) {
	if h.fileService == nil {
		response.BadRequest(c, "file storage unavailable")
		return
	}

	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	file, header, err := c.Request.FormFile("icon")
	if err != nil {
		response.BadRequest(c, "icon file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	uploaded, err := h.fileService.Upload(c.Request.Context(), userID, header.Filename, file, header.Size, contentType)
	if err != nil {
		response.BadRequest(c, fmt.Sprintf("upload failed: %v", err))
		return
	}

	workspace, err := h.workspaceService.Update(c.Request.Context(), workspaceID, domain.WorkspaceUpdate{
		IconURL: &uploaded.URL,
	}, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToWorkspace(workspaceID, "workspace_updated", workspace, "")
	response.Success(c, workspace)
}

// RegisterRoutes регистрирует маршруты
func (h *WorkspaceHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, channelHandler *ChannelHandler, roleHandler *WorkspaceRoleHandler, categoryHandler *ChannelCategoryHandler) {
	workspaces := r.Group("/workspaces", authMiddleware)
	{
		workspaces.POST("", h.Create)
		workspaces.GET("", h.GetUserWorkspaces)
		workspaces.GET("/:id", h.GetByID)
		workspaces.PATCH("/:id", h.Update)
		workspaces.DELETE("/:id", h.Delete)
		workspaces.POST("/:id/icon", h.UploadIcon)
		workspaces.GET("/:id/members", h.GetMembers)
		workspaces.POST("/:id/members", h.AddMember)
		workspaces.PATCH("/:id/members/:userId", h.UpdateMember)
		workspaces.DELETE("/:id/members/:userId", h.RemoveMember)
		workspaces.GET("/:id/channels", channelHandler.GetByWorkspaceID)

		// Категории каналов
		categoryHandler.RegisterRoutes(workspaces)

		// Роли воркспейса
		roleHandler.RegisterRoutes(workspaces.Group("/:id/roles"))
		roleHandler.RegisterMemberRoleRoutes(workspaces.Group("/:id/members/:userId"))
	}
}

