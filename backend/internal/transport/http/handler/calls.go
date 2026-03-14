package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type CallsHandler struct {
	lkService      *service.LiveKitService
	authService    *service.AuthService
	channelService *service.ChannelService
	wsHub          *ws.Hub
}

func NewCallsHandler(
	lkService *service.LiveKitService,
	authService *service.AuthService,
	channelService *service.ChannelService,
	wsHub *ws.Hub,
) *CallsHandler {
	return &CallsHandler{
		lkService:      lkService,
		authService:    authService,
		channelService: channelService,
		wsHub:          wsHub,
	}
}

// JoinCall POST /api/v1/calls/join
func (h *CallsHandler) JoinCall(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		ChannelID string `json:"channel_id" binding:"required"`
		RoomName  string `json:"room_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Доступ проверяется внутри GetByID → getAccessibleChannel.
	// Для DM там вызывается IsMember, для voice-каналов — workspace-роль.
	_, err := h.channelService.GetByID(c.Request.Context(), body.ChannelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	roomName := body.RoomName
	if roomName == "" {
		roomName = h.lkService.CreateRoomName(body.ChannelID)
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, service.ErrUserNotFound)
		return
	}

	token, err := h.lkService.CreateToken(c.Request.Context(), roomName, userID, user.DisplayName, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create call token"})
		return
	}

	response.Success(c, token)
}

// SignalCall POST /api/v1/calls/signal — отправляет WS-сигнал участнику DM о входящем/завершённом звонке
func (h *CallsHandler) SignalCall(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		ChannelID string `json:"channel_id" binding:"required,uuid"`
		Signal    string `json:"signal"     binding:"required,oneof=started ended"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// GetByID проверяет доступ — для DM через IsMember
	channel, err := h.channelService.GetByID(c.Request.Context(), body.ChannelID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if channel.Type != domain.ChannelTypeDM {
		response.BadRequest(c, "signal is only supported for DM channels")
		return
	}

	members, err := h.channelService.GetChannelMembers(c.Request.Context(), body.ChannelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get members"})
		return
	}

	caller, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, service.ErrUserNotFound)
		return
	}

	// Отправляем событие только собеседнику (не себе)
	recipientIDs := make([]string, 0, 1)
	for _, m := range members {
		if m.UserID != userID {
			recipientIDs = append(recipientIDs, m.UserID)
		}
	}

	payload := gin.H{
		"channel_id":    body.ChannelID,
		"caller_id":     userID,
		"caller_name":   caller.DisplayName,
		"caller_avatar": caller.AvatarURL,
	}

	h.wsHub.BroadcastToUsers(recipientIDs, "dm_call_"+body.Signal, payload)

	c.JSON(http.StatusNoContent, nil)
}
