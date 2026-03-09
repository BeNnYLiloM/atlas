package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

type CallsHandler struct {
	lkService *service.LiveKitService
	authService *service.AuthService
}

func NewCallsHandler(lkService *service.LiveKitService, authService *service.AuthService) *CallsHandler {
	return &CallsHandler{lkService: lkService, authService: authService}
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

	roomName := body.RoomName
	if roomName == "" {
		roomName = h.lkService.CreateRoomName(body.ChannelID)
	}

	// Получаем пользователя для display_name
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
