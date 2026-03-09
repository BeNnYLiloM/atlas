package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type WorkspaceRoleHandler struct {
	roleService    *service.WorkspaceRoleService
	channelService *service.ChannelService
	wsHub          *ws.Hub
}

func NewWorkspaceRoleHandler(roleService *service.WorkspaceRoleService, channelService *service.ChannelService, wsHub *ws.Hub) *WorkspaceRoleHandler {
	return &WorkspaceRoleHandler{
		roleService:    roleService,
		channelService: channelService,
		wsHub:          wsHub,
	}
}

// List — GET /workspaces/:id/roles
func (h *WorkspaceRoleHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	roles, err := h.roleService.List(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, roles)
}

// Create — POST /workspaces/:id/roles
func (h *WorkspaceRoleHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.WorkspaceRoleCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	role, err := h.roleService.Create(c.Request.Context(), workspaceID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, role)
}

// Update — PATCH /workspaces/:id/roles/:roleId
func (h *WorkspaceRoleHandler) Update(c *gin.Context) {
	workspaceID := c.Param("id")
	roleID := c.Param("roleId")
	userID := middleware.GetUserID(c)

	var input domain.WorkspaceRoleUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	role, err := h.roleService.Update(c.Request.Context(), workspaceID, roleID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, role)
}

// UpdateEveryone — PATCH /workspaces/:id/roles/everyone
func (h *WorkspaceRoleHandler) UpdateEveryone(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var perms domain.RolePermissions
	if err := c.ShouldBindJSON(&perms); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	role, err := h.roleService.UpdateEveryonePermissions(c.Request.Context(), workspaceID, perms, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, role)
}

// Delete — DELETE /workspaces/:id/roles/:roleId
func (h *WorkspaceRoleHandler) Delete(c *gin.Context) {
	workspaceID := c.Param("id")
	roleID := c.Param("roleId")
	userID := middleware.GetUserID(c)

	if err := h.roleService.Delete(c.Request.Context(), workspaceID, roleID, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// AssignRole — POST /workspaces/:id/members/:userId/roles
func (h *WorkspaceRoleHandler) AssignRole(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	userID := middleware.GetUserID(c)

	var body struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.roleService.AssignRole(c.Request.Context(), workspaceID, targetUserID, body.RoleID, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Получаем каналы пока контекст ещё жив, затем рассылаем в горутине
	channels, _ := h.channelService.GetChannelsByRole(c.Request.Context(), body.RoleID)
	go func() {
		for _, ch := range channels {
			h.wsHub.SendToUser(targetUserID, "channel_created", ch)
		}
		// Уведомляем всех участников воркспейса об изменении роли
		h.wsHub.BroadcastToWorkspace(workspaceID, "member_updated", map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      targetUserID,
		}, "")
	}()

	response.NoContent(c)
}

// RevokeRole — DELETE /workspaces/:id/members/:userId/roles/:roleId
func (h *WorkspaceRoleHandler) RevokeRole(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	roleID := c.Param("roleId")
	userID := middleware.GetUserID(c)

	// Собираем каналы ДО отзыва роли пока контекст жив
	channels, _ := h.channelService.GetChannelsByRole(c.Request.Context(), roleID)

	if err := h.roleService.RevokeRole(c.Request.Context(), workspaceID, targetUserID, roleID, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Рассылаем только приватные каналы — публичные остаются видны без роли
	go func() {
		for _, ch := range channels {
			if ch.IsPrivate {
				h.wsHub.SendToUser(targetUserID, "channel_deleted", map[string]string{
					"workspace_id": ch.WorkspaceID,
					"channel_id":   ch.ID,
				})
			}
		}
		// Уведомляем всех участников воркспейса об изменении роли
		h.wsHub.BroadcastToWorkspace(workspaceID, "member_updated", map[string]interface{}{
			"workspace_id": workspaceID,
			"user_id":      targetUserID,
		}, "")
	}()

	response.NoContent(c)
}

// GetMemberRoles — GET /workspaces/:id/members/:userId/roles
func (h *WorkspaceRoleHandler) GetMemberRoles(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("userId")
	userID := middleware.GetUserID(c)

	roles, err := h.roleService.GetMemberRoles(c.Request.Context(), workspaceID, targetUserID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, roles)
}

// RegisterRoutes регистрирует маршруты ролей в рамках workspace handler
func (h *WorkspaceRoleHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", h.List)
	r.POST("", h.Create)
	r.PATCH("/everyone", h.UpdateEveryone)
	r.PATCH("/:roleId", h.Update)
	r.DELETE("/:roleId", h.Delete)
}

func (h *WorkspaceRoleHandler) RegisterMemberRoleRoutes(r *gin.RouterGroup) {
	r.POST("/roles", h.AssignRole)
	r.GET("/roles", h.GetMemberRoles)
	r.DELETE("/roles/:roleId", h.RevokeRole)
}
