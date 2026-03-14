package handler

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type MessageHandler struct {
	messageService *service.MessageService
	channelService *service.ChannelService
	projectService *service.ProjectService
	wsHub          *ws.Hub
}

func NewMessageHandler(messageService *service.MessageService, channelService *service.ChannelService, projectService *service.ProjectService, wsHub *ws.Hub) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		channelService: channelService,
		projectService: projectService,
		wsHub:          wsHub,
	}
}

// Create создает новое сообщение
func (h *MessageHandler) Create(c *gin.Context) {
	var input domain.MessageCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	message, err := h.messageService.Create(c.Request.Context(), input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Получаем канал чтобы знать workspaceID и projectID
	ctx := c.Request.Context()
	channel, err := h.channelService.GetByID(ctx, message.ChannelID, userID)
	if err == nil && channel != nil {
		// Для проектных каналов — рассылаем только участникам проекта + view_all_projects
		broadcastFn := func(event string, data interface{}) {
			if channel.ProjectID != nil {
				recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(ctx, *channel.ProjectID, channel.WorkspaceID)
				filtered := make([]string, 0, len(recipientIDs))
				for _, id := range recipientIDs {
					if id != userID {
						filtered = append(filtered, id)
					}
				}
				h.wsHub.BroadcastToUsers(filtered, event, data)
			} else {
				h.wsHub.BroadcastToWorkspace(channel.WorkspaceID, event, data, userID)
			}
		}

		if message.ParentID != nil {
			broadcastFn("thread_reply", map[string]interface{}{
				"channel_id": message.ChannelID,
				"parent_id":  *message.ParentID,
				"message":    message,
			})
		} else {
			broadcastFn("message", map[string]interface{}{
				"channel_id": message.ChannelID,
				"message":    message,
			})
		}
	}

	response.Created(c, message)
}

// GetByChannelID возвращает сообщения канала
func (h *MessageHandler) GetByChannelID(c *gin.Context) {
	channelID := c.Param("id")
	userID := middleware.GetUserID(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.messageService.GetByChannelID(c.Request.Context(), channelID, userID, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, messages)
}

// GetThreadMessages возвращает сообщения треда
func (h *MessageHandler) GetThreadMessages(c *gin.Context) {
	messageID := c.Param("id")
	userID := middleware.GetUserID(c)

	messages, err := h.messageService.GetThreadMessages(c.Request.Context(), messageID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, messages)
}

// Update обновляет сообщение
func (h *MessageHandler) Update(c *gin.Context) {
	messageID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.MessageUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	message, err := h.messageService.Update(c.Request.Context(), messageID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Получаем канал для workspaceID
	ctx := c.Request.Context()
	channel, err := h.channelService.GetByID(ctx, message.ChannelID, userID)
	if err == nil && channel != nil {
		// Broadcast обновления сообщения в workspace (исключаем редактора)
		h.wsHub.BroadcastToWorkspace(
			channel.WorkspaceID,
			"message_updated",
			map[string]interface{}{
				"channel_id": message.ChannelID,
				"message":    message,
			},
			userID, // Исключаем редактора
		)
	}

	response.Success(c, message)
}

// Delete удаляет сообщение
func (h *MessageHandler) Delete(c *gin.Context) {
	messageID := c.Param("id")
	userID := middleware.GetUserID(c)

	channelID, err := h.messageService.Delete(c.Request.Context(), messageID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Получаем канал для workspaceID
	ctx := c.Request.Context()
	channel, err := h.channelService.GetByID(ctx, channelID, userID)
	if err == nil && channel != nil {
		// Broadcast удаления сообщения в workspace (исключаем удалившего)
		h.wsHub.BroadcastToWorkspace(
			channel.WorkspaceID,
			"message_deleted",
			map[string]interface{}{
				"channel_id": channelID,
				"message_id": messageID,
			},
			userID, // Исключаем удалившего
		)
	}

	response.NoContent(c)
}

// MarkThreadAsRead отмечает тред прочитанным
func (h *MessageHandler) MarkThreadAsRead(c *gin.Context) {
	parentMessageID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input struct {
		MessageID *string `json:"message_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.messageService.MarkThreadAsRead(c.Request.Context(), parentMessageID, userID, input.MessageID)
	if err != nil {
		log.Printf("Error marking thread as read: %v", err)
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// GetThreadUnreadCount возвращает количество непрочитанных сообщений в треде
func (h *MessageHandler) GetThreadUnreadCount(c *gin.Context) {
	parentMessageID := c.Param("id")
	userID := middleware.GetUserID(c)

	count, err := h.messageService.GetThreadUnreadCount(c.Request.Context(), parentMessageID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]interface{}{
		"parent_id":    parentMessageID,
		"unread_count": count,
	})
}

// RegisterRoutes регистрирует маршруты
func (h *MessageHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	messages := r.Group("/messages", authMiddleware)
	{
		messages.POST("", h.Create)
		messages.GET("/:id/thread", h.GetThreadMessages)
		messages.POST("/:id/thread/read", h.MarkThreadAsRead)
		messages.GET("/:id/thread/unread", h.GetThreadUnreadCount)
		messages.PUT("/:id", h.Update)
		messages.DELETE("/:id", h.Delete)
	}

	// Сообщения канала - регистрируется в ChannelHandler
}

