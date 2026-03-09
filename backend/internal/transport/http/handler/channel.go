package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type ChannelHandler struct {
	channelService *service.ChannelService
	wsHub          *ws.Hub
}

func NewChannelHandler(channelService *service.ChannelService, wsHub *ws.Hub) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
		wsHub:          wsHub,
	}
}

// Create создает новый канал
func (h *ChannelHandler) Create(c *gin.Context) {
	var input domain.ChannelCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	channel, err := h.channelService.Create(c.Request.Context(), input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Для публичных каналов — broadcast всем, для приватных — только тем у кого есть доступ
	if !channel.IsPrivate {
		h.wsHub.BroadcastToWorkspace(channel.WorkspaceID, "channel_created", channel, userID)
	} else {
		accessIDs, err := h.channelService.GetAccessibleUserIDs(c.Request.Context(), channel)
		if err == nil {
			// Исключаем создателя из списка (он уже получил ответ)
			filtered := make([]string, 0, len(accessIDs))
			for _, id := range accessIDs {
				if id != userID {
					filtered = append(filtered, id)
				}
			}
			h.wsHub.BroadcastToUsers(filtered, "channel_created", channel)
		}
	}

	response.Created(c, channel)
}

// GetByID возвращает канал по ID
func (h *ChannelHandler) GetByID(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, channel)
}

// GetByWorkspaceID возвращает каналы воркспейса с unread counts
func (h *ChannelHandler) GetByWorkspaceID(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	channels, err := h.channelService.GetByWorkspaceIDWithUnread(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, channels)
}

// Update обновляет настройки канала (name, topic, is_private, slowmode_seconds)
func (h *ChannelHandler) Update(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.ChannelUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Получаем состояние ДО обновления чтобы понять изменился ли is_private
	oldChannel, _ := h.channelService.GetByID(c.Request.Context(), channelID, userID)

	channel, err := h.channelService.Update(c.Request.Context(), channelID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	wasPrivate := oldChannel != nil && oldChannel.IsPrivate
	isPrivate := channel.IsPrivate
	privacyChanged := input.IsPrivate != nil && wasPrivate != isPrivate

	if privacyChanged {
		// Получаем всех участников воркспейса
		allUserIDs, _ := h.channelService.GetAllWorkspaceUserIDs(c.Request.Context(), channel.WorkspaceID)

		if !isPrivate {
			// приватный → публичный: channel_created всем кто раньше не имел доступа + channel_updated тем кто имел
			accessIDs, _ := h.channelService.GetAccessibleUserIDs(c.Request.Context(), oldChannel)
			hadAccess := make(map[string]bool, len(accessIDs))
			for _, id := range accessIDs {
				hadAccess[id] = true
			}
			for _, id := range allUserIDs {
				if id == userID {
					continue
				}
				if hadAccess[id] {
					h.wsHub.SendToUser(id, "channel_updated", channel)
				} else {
					h.wsHub.SendToUser(id, "channel_created", channel)
				}
			}
		} else {
			// публичный → приватный: channel_deleted тем у кого нет доступа, channel_updated тем у кого есть
			accessIDs, _ := h.channelService.GetAccessibleUserIDs(c.Request.Context(), channel)
			hasAccess := make(map[string]bool, len(accessIDs))
			for _, id := range accessIDs {
				hasAccess[id] = true
			}
			for _, id := range allUserIDs {
				if id == userID {
					continue
				}
				if hasAccess[id] {
					h.wsHub.SendToUser(id, "channel_updated", channel)
				} else {
					h.wsHub.SendToUser(id, "channel_deleted", map[string]string{
						"workspace_id": channel.WorkspaceID,
						"channel_id":   channelID,
					})
				}
			}
		}
	} else if !isPrivate {
		// Публичный без смены приватности — обычный broadcast
		h.wsHub.BroadcastToWorkspace(channel.WorkspaceID, "channel_updated", channel, userID)
	} else {
		// Приватный без смены приватности — только тем у кого есть доступ
		accessIDs, err := h.channelService.GetAccessibleUserIDs(c.Request.Context(), channel)
		if err == nil {
			filtered := make([]string, 0, len(accessIDs))
			for _, id := range accessIDs {
				if id != userID {
					filtered = append(filtered, id)
				}
			}
			h.wsHub.BroadcastToUsers(filtered, "channel_updated", channel)
		}
	}

	response.Success(c, channel)
}

// GetMembers возвращает участников канала
func (h *ChannelHandler) GetMembers(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	members, err := h.channelService.GetChannelMembers(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, members)
}

// AddMember добавляет участника в канал
func (h *ChannelHandler) AddMember(c *gin.Context) {
	channelID := c.Param("id")
	actorID := middleware.GetUserID(c)

	var input struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.channelService.AddChannelMember(c.Request.Context(), channelID, input.UserID, actorID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// RemoveMember удаляет участника из канала
func (h *ChannelHandler) RemoveMember(c *gin.Context) {
	channelID := c.Param("id")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	if err := h.channelService.RemoveChannelMember(c.Request.Context(), channelID, targetUserID, actorID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// UpdateNotifications обновляет уровень уведомлений текущего пользователя
func (h *ChannelHandler) UpdateNotifications(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.UpdateNotificationsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.channelService.UpdateNotifications(c.Request.Context(), channelID, userID, input.Level); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// Delete удаляет канал
func (h *ChannelHandler) Delete(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	// Сначала получаем канал чтобы знать workspaceID
	channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	err = h.channelService.Delete(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Broadcast удаления канала в workspace (исключаем инициатора)
	h.wsHub.BroadcastToWorkspace(
		channel.WorkspaceID,
		"channel_deleted",
		map[string]string{
			"workspace_id": channel.WorkspaceID,
			"channel_id":   channelID,
		},
		userID, // Исключаем инициатора из broadcast
	)

	response.NoContent(c)
}

// MarkAsRead отмечает канал прочитанным
func (h *ChannelHandler) MarkAsRead(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.MarkAsReadInput
	// Если не передали body - просто отмечаем канал прочитанным
	_ = c.ShouldBindJSON(&input)

	err := h.channelService.MarkAsRead(c.Request.Context(), channelID, userID, input.MessageID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Broadcast read_state_update в workspace
	channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID)
	if err == nil && channel != nil {
		h.wsHub.BroadcastToWorkspace(
			channel.WorkspaceID,
			"read_state_update",
			map[string]interface{}{
				"channel_id": channelID,
				"user_id":    userID,
			},
			userID, // Исключаем текущего пользователя
		)
	}

	response.NoContent(c)
}

// GetPermissions возвращает права доступа канала
func (h *ChannelHandler) GetPermissions(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	perms, err := h.channelService.GetPermissions(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, perms)
}

// AddRolePermission добавляет роль в доступ к каналу
func (h *ChannelHandler) AddRolePermission(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.AddRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.channelService.AddRolePermission(c.Request.Context(), channelID, input, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Уведомляем всех участников с этой ролью о появлении канала
	if channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID); err == nil {
		if roleUserIDs, err := h.channelService.GetRoleUserIDs(c.Request.Context(), input.RoleID); err == nil {
			for _, uid := range roleUserIDs {
				h.wsHub.SendToUser(uid, "channel_created", channel)
			}
		}
	}

	response.NoContent(c)
}

// RemoveRolePermission удаляет роль из доступа к каналу
func (h *ChannelHandler) RemoveRolePermission(c *gin.Context) {
	channelID := c.Param("id")
	roleID := c.Param("roleId")
	userID := middleware.GetUserID(c)

	// Получаем участников с этой ролью ДО удаления
	roleUserIDs, _ := h.channelService.GetRoleUserIDs(c.Request.Context(), roleID)

	if err := h.channelService.RemoveRolePermission(c.Request.Context(), channelID, roleID, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Уведомляем участников с этой ролью об исчезновении канала
	if channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID); err == nil {
		for _, uid := range roleUserIDs {
			h.wsHub.SendToUser(uid, "channel_deleted", map[string]string{
				"workspace_id": channel.WorkspaceID,
				"channel_id":   channelID,
			})
		}
	}

	response.NoContent(c)
}

// AddUserPermission добавляет участника в доступ к каналу
func (h *ChannelHandler) AddUserPermission(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.AddUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.channelService.AddUserPermission(c.Request.Context(), channelID, input, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Уведомляем добавленного участника о канале
	if channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID); err == nil {
		h.wsHub.SendToUser(input.UserID, "channel_created", channel)
	}

	response.NoContent(c)
}

// RemoveUserPermission удаляет участника из доступа к каналу
func (h *ChannelHandler) RemoveUserPermission(c *gin.Context) {
	channelID := c.Param("id")
	targetUserID := c.Param("userId")
	userID := middleware.GetUserID(c)

	if err := h.channelService.RemoveUserPermission(c.Request.Context(), channelID, targetUserID, userID); err != nil {
		response.Error(c, err)
		return
	}

	// Уведомляем удалённого участника об исчезновении канала
	if channel, err := h.channelService.GetByID(c.Request.Context(), channelID, userID); err == nil {
		h.wsHub.SendToUser(targetUserID, "channel_deleted", map[string]string{
			"workspace_id": channel.WorkspaceID,
			"channel_id":   channelID,
		})
	}

	response.NoContent(c)
}

// CheckCanWrite проверяет может ли текущий пользователь писать в канал
func (h *ChannelHandler) CheckCanWrite(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	canWrite, err := h.channelService.CanUserWrite(c.Request.Context(), channelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, map[string]bool{"can_write": canWrite})
}

// RegisterRoutes регистрирует маршруты
func (h *ChannelHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	channels := r.Group("/channels", authMiddleware)
	{
		channels.POST("", h.Create)
		channels.GET("/:id", h.GetByID)
		channels.PATCH("/:id", h.Update)
		channels.DELETE("/:id", h.Delete)
		channels.POST("/:id/read", h.MarkAsRead)
		channels.PATCH("/:id/notifications", h.UpdateNotifications)
	}
}

// RegisterWithMessages регистрирует маршруты с поддержкой сообщений
func (h *ChannelHandler) RegisterWithMessages(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, messageHandler *MessageHandler) {
	channels := r.Group("/channels", authMiddleware)
	{
		channels.POST("", h.Create)
		channels.GET("/:id", h.GetByID)
		channels.PATCH("/:id", h.Update)
		channels.DELETE("/:id", h.Delete)
		channels.POST("/:id/read", h.MarkAsRead)
		channels.GET("/:id/messages", messageHandler.GetByChannelID)
		channels.PATCH("/:id/notifications", h.UpdateNotifications)
		channels.GET("/:id/members", h.GetMembers)
		channels.POST("/:id/members", h.AddMember)
		channels.DELETE("/:id/members/:userId", h.RemoveMember)
		// Права доступа
		channels.GET("/:id/permissions", h.GetPermissions)
		channels.GET("/:id/can-write", h.CheckCanWrite)
		channels.POST("/:id/permissions/roles", h.AddRolePermission)
		channels.DELETE("/:id/permissions/roles/:roleId", h.RemoveRolePermission)
		channels.POST("/:id/permissions/users", h.AddUserPermission)
		channels.DELETE("/:id/permissions/users/:userId", h.RemoveUserPermission)
	}
}

