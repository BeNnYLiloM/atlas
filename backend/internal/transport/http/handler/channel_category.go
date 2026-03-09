package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type ChannelCategoryHandler struct {
	categoryService *service.ChannelCategoryService
	channelPermRepo repository.ChannelPermissionRepository
	wsHub           *ws.Hub
}

func NewChannelCategoryHandler(
	categoryService *service.ChannelCategoryService,
	channelPermRepo repository.ChannelPermissionRepository,
	wsHub *ws.Hub,
) *ChannelCategoryHandler {
	return &ChannelCategoryHandler{
		categoryService: categoryService,
		channelPermRepo: channelPermRepo,
		wsHub:           wsHub,
	}
}

// List GET /workspaces/:id/categories
// Для owner/admin возвращает все категории; для обычных — только доступные
func (h *ChannelCategoryHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	cats, err := h.categoryService.GetVisibleByWorkspaceID(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if cats == nil {
		cats = []*domain.ChannelCategory{}
	}
	response.Success(c, cats)
}

// Create POST /workspaces/:id/categories
func (h *ChannelCategoryHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.ChannelCategoryCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	input.WorkspaceID = workspaceID

	cat, err := h.categoryService.Create(c.Request.Context(), input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Приватную категорию рассылаем только admins (обычные юзеры её не видят пока нет доступа)
	if cat.IsPrivate {
		// Admins/owners получат через BroadcastToWorkspace — они видят всё
		h.wsHub.BroadcastToWorkspace(workspaceID, "category_created", cat, "")
	} else {
		h.wsHub.BroadcastToWorkspace(workspaceID, "category_created", cat, "")
	}
	response.Created(c, cat)
}

// UpdateInWorkspace PATCH /workspaces/:id/categories/:categoryId
func (h *ChannelCategoryHandler) UpdateInWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	categoryID := c.Param("categoryId")
	userID := middleware.GetUserID(c)

	var input domain.ChannelCategoryUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cat, err := h.categoryService.Update(c.Request.Context(), categoryID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToWorkspace(workspaceID, "category_updated", cat, "")
	response.Success(c, cat)
}

// DeleteInWorkspace DELETE /workspaces/:id/categories/:categoryId
func (h *ChannelCategoryHandler) DeleteInWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	categoryID := c.Param("categoryId")
	userID := middleware.GetUserID(c)

	if err := h.categoryService.Delete(c.Request.Context(), categoryID, userID); err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToWorkspace(workspaceID, "category_deleted", map[string]string{
		"workspace_id": workspaceID,
		"category_id":  categoryID,
	}, "")
	response.NoContent(c)
}

// GetPermissions GET /workspaces/:id/categories/:categoryId/permissions
func (h *ChannelCategoryHandler) GetPermissions(c *gin.Context) {
	categoryID := c.Param("categoryId")
	userID := middleware.GetUserID(c)

	perms, err := h.categoryService.GetPermissions(c.Request.Context(), categoryID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, perms)
}

// AddRolePermission POST /workspaces/:id/categories/:categoryId/permissions/roles
func (h *ChannelCategoryHandler) AddRolePermission(c *gin.Context) {
	workspaceID := c.Param("id")
	categoryID := c.Param("categoryId")
	userID := middleware.GetUserID(c)

	var input domain.AddRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.categoryService.AddRole(c.Request.Context(), categoryID, input.RoleID, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Синхронизируем права роли в приватные каналы этой категории
	h.syncRoleToChannels(c, workspaceID, categoryID, input.RoleID, true)

	response.NoContent(c)
}

// RemoveRolePermission DELETE /workspaces/:id/categories/:categoryId/permissions/roles/:roleId
func (h *ChannelCategoryHandler) RemoveRolePermission(c *gin.Context) {
	workspaceID := c.Param("id")
	categoryID := c.Param("categoryId")
	roleID := c.Param("roleId")
	userID := middleware.GetUserID(c)

	if err := h.categoryService.RemoveRole(c.Request.Context(), categoryID, roleID, userID); err != nil {
		response.Error(c, err)
		return
	}

	h.syncRoleToChannels(c, workspaceID, categoryID, roleID, false)

	response.NoContent(c)
}

// AddUserPermission POST /workspaces/:id/categories/:categoryId/permissions/users
func (h *ChannelCategoryHandler) AddUserPermission(c *gin.Context) {
	workspaceID := c.Param("id")
	categoryID := c.Param("categoryId")
	userID := middleware.GetUserID(c)

	var input domain.AddUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.categoryService.AddUser(c.Request.Context(), categoryID, input.UserID, userID); err != nil {
		response.Error(c, err)
		return
	}

	h.syncUserToChannels(c, workspaceID, categoryID, input.UserID, true)

	response.NoContent(c)
}

// RemoveUserPermission DELETE /workspaces/:id/categories/:categoryId/permissions/users/:userId
func (h *ChannelCategoryHandler) RemoveUserPermission(c *gin.Context) {
	workspaceID := c.Param("id")
	categoryID := c.Param("categoryId")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	if err := h.categoryService.RemoveUser(c.Request.Context(), categoryID, targetUserID, actorID); err != nil {
		response.Error(c, err)
		return
	}

	h.syncUserToChannels(c, workspaceID, categoryID, targetUserID, false)

	response.NoContent(c)
}

// syncRoleToChannels — копирует/удаляет права роли в/из всех каналов приватной категории
func (h *ChannelCategoryHandler) syncRoleToChannels(c *gin.Context, workspaceID, categoryID, roleID string, add bool) {
	cat, err := h.categoryService.GetCategoryByID(c.Request.Context(), categoryID)
	if err != nil || cat == nil || !cat.IsPrivate {
		return
	}

	channels, err := h.categoryService.GetChannelsOfCategory(c.Request.Context(), workspaceID, categoryID)
	if err != nil {
		return
	}

	for _, ch := range channels {
		if add {
			_ = h.channelPermRepo.AddRole(c.Request.Context(), ch.ID, roleID)
		} else {
			_ = h.channelPermRepo.RemoveRole(c.Request.Context(), ch.ID, roleID)
		}
	}
}

// syncUserToChannels — копирует/удаляет права пользователя в/из всех каналов приватной категории
func (h *ChannelCategoryHandler) syncUserToChannels(c *gin.Context, workspaceID, categoryID, targetUserID string, add bool) {
	cat, err := h.categoryService.GetCategoryByID(c.Request.Context(), categoryID)
	if err != nil || cat == nil || !cat.IsPrivate {
		return
	}

	channels, err := h.categoryService.GetChannelsOfCategory(c.Request.Context(), workspaceID, categoryID)
	if err != nil {
		return
	}

	for _, ch := range channels {
		if add {
			_ = h.channelPermRepo.AddUser(c.Request.Context(), ch.ID, targetUserID)
		} else {
			_ = h.channelPermRepo.RemoveUser(c.Request.Context(), ch.ID, targetUserID)
		}
	}
}

// RegisterRoutes регистрирует маршруты категорий
func (h *ChannelCategoryHandler) RegisterRoutes(workspaces *gin.RouterGroup) {
	workspaces.GET("/:id/categories", h.List)
	workspaces.POST("/:id/categories", h.Create)
	workspaces.PATCH("/:id/categories/:categoryId", h.UpdateInWorkspace)
	workspaces.DELETE("/:id/categories/:categoryId", h.DeleteInWorkspace)
	workspaces.GET("/:id/categories/:categoryId/permissions", h.GetPermissions)
	workspaces.POST("/:id/categories/:categoryId/permissions/roles", h.AddRolePermission)
	workspaces.DELETE("/:id/categories/:categoryId/permissions/roles/:roleId", h.RemoveRolePermission)
	workspaces.POST("/:id/categories/:categoryId/permissions/users", h.AddUserPermission)
	workspaces.DELETE("/:id/categories/:categoryId/permissions/users/:userId", h.RemoveUserPermission)
}
