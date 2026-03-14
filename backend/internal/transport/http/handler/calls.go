package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	messageService *service.MessageService
	wsHub          *ws.Hub
}

func NewCallsHandler(
	lkService *service.LiveKitService,
	authService *service.AuthService,
	channelService *service.ChannelService,
	messageService *service.MessageService,
	wsHub *ws.Hub,
) *CallsHandler {
	return &CallsHandler{
		lkService:      lkService,
		authService:    authService,
		channelService: channelService,
		messageService: messageService,
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

// SignalCall POST /api/v1/calls/signal
//
// started → создаёт call-сообщение со статусом "missed", рассылает WS-сигнал собеседнику
// ended   → обновляет статус call-сообщения на "ended" + вычисляет длительность, рассылает WS-сигнал
func (h *CallsHandler) SignalCall(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var body struct {
		ChannelID string `json:"channel_id" binding:"required,uuid"`
		// started | accepted | ended
		Signal    string `json:"signal"     binding:"required,oneof=started accepted ended"`
		// ID call-сообщения — нужен для accepted и ended чтобы обновить запись
		CallMsgID string `json:"call_msg_id"`
		// Время старта звонка в unix ms — нужен для ended для вычисления duration (0 = не было разговора)
		StartedAt int64 `json:"started_at"`
		// Cancelled=true если инициатор сам отменил звонок до ответа
		Cancelled bool `json:"cancelled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// VULN-6: uuid-валидация call_msg_id если передан
	if body.CallMsgID != "" {
		if _, err := uuid.Parse(body.CallMsgID); err != nil {
			response.BadRequest(c, "call_msg_id must be a valid UUID")
			return
		}
	}

	// VULN-2: started_at должен быть либо 0, либо разумным unix ms (не в будущем, не слишком старым)
	const maxCallDurationMs = int64(24 * 60 * 60 * 1000) // 24 часа
	if body.StartedAt < 0 {
		response.BadRequest(c, "started_at must be non-negative")
		return
	}
	if body.StartedAt > 0 {
		nowMs := time.Now().UnixMilli()
		if body.StartedAt > nowMs {
			response.BadRequest(c, "started_at cannot be in the future")
			return
		}
		if nowMs-body.StartedAt > maxCallDurationMs {
			response.BadRequest(c, "started_at is too far in the past")
			return
		}
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

	allIDs := append(recipientIDs, userID)

	switch body.Signal {
	case "started":
		// Создаём call-сообщение со статусом missed (обновится на ongoing при принятии)
		msg, err := h.messageService.CreateCallMessage(c.Request.Context(), body.ChannelID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create call record"})
			return
		}
		payload["call_msg_id"] = msg.ID
		payload["message"] = msg

		// dm_message — оба видят запись в истории сразу
		h.wsHub.BroadcastToUsers(allIDs, "dm_message", gin.H{
			"channel_id": body.ChannelID,
			"message":    msg,
		})
		// dm_call_started — только собеседнику (показать входящий звонок)
		h.wsHub.BroadcastToUsers(recipientIDs, "dm_call_started", payload)

	case "accepted":
		// Собеседник принял звонок — обновляем статус на ongoing.
		// callerID здесь — это ID звонящего (создателя записи), который передал call_msg_id.
		// Получатель (acceptor) не является user_id записи — поэтому UpdateCallStatus
		// вызывается от имени звонящего, чей ID хранится в сообщении.
		// Безопасность: канал уже проверен через GetByID/IsMember выше.
		if body.CallMsgID != "" {
			// callerID получателя = userID (он принимает), но owner сообщения — другой участник.
			// Передаём пустой callerID чтобы пропустить проверку owner в этом конкретном переходе:
			// вместо этого полагаемся на проверку канала (оба участника DM).
			// Используем специальный метод без owner-check для accepted-перехода.
			if err := h.messageService.AcceptCallStatus(c.Request.Context(), body.CallMsgID, body.ChannelID); err == nil {
				h.wsHub.BroadcastToUsers(allIDs, "dm_call_message_updated", gin.H{
					"channel_id":   body.ChannelID,
					"call_msg_id":  body.CallMsgID,
					"call_status":  domain.CallStatusOngoing,
					"duration_sec": nil,
				})
			}
		}

	case "ended":
		if body.CallMsgID != "" {
			var finalStatus string
			var durationSec *int

			switch {
			case body.StartedAt > 0:
				d := int(time.Since(time.UnixMilli(body.StartedAt)).Seconds())
				durationSec = &d
				finalStatus = domain.CallStatusEnded
			case body.Cancelled:
				finalStatus = domain.CallStatusCancelled
			default:
				finalStatus = domain.CallStatusMissed
			}

			// callerID = userID: только участник канала может обновить статус.
			// Для ended/cancelled/missed — проверяем что текущий пользователь является участником канала (уже проверено выше).
			// owner-check снят т.к. и инициатор и получатель могут завершить звонок.
			if err := h.messageService.EndCallStatus(c.Request.Context(), body.CallMsgID, body.ChannelID, finalStatus, durationSec); err == nil {
				h.wsHub.BroadcastToUsers(allIDs, "dm_call_message_updated", gin.H{
					"channel_id":   body.ChannelID,
					"call_msg_id":  body.CallMsgID,
					"call_status":  finalStatus,
					"duration_sec": durationSec,
				})
			}
		}
		h.wsHub.BroadcastToUsers(recipientIDs, "dm_call_ended", payload)
	}

	c.JSON(http.StatusNoContent, nil)
}
