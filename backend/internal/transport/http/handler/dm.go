package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

type DMHandler struct {
	dmService *service.DMService
}

func NewDMHandler(dmService *service.DMService) *DMHandler {
	return &DMHandler{dmService: dmService}
}

type openDMRequest struct {
	WorkspaceID  string `json:"workspace_id"   binding:"required,uuid"`
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
}

// Open godoc
// @Summary Открыть или создать DM-канал
// @Tags dm
// @Accept json
// @Produce json
// @Param body body openDMRequest true "target_user_id + workspace_id"
// @Success 200 {object} response.SuccessResponse
// @Router /dm [post]
func (h *DMHandler) Open(c *gin.Context) {
	var req openDMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	channel, err := h.dmService.GetOrCreateDM(c.Request.Context(), req.WorkspaceID, userID, req.TargetUserID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, channel)
}

// List godoc
// @Summary Список DM-диалогов пользователя
// @Tags dm
// @Produce json
// @Param workspace_id query string true "ID воркспейса"
// @Success 200 {object} response.SuccessResponse
// @Router /dm [get]
func (h *DMHandler) List(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		response.BadRequest(c, "workspace_id must be a valid UUID")
		return
	}

	userID := middleware.GetUserID(c)
	dms, err := h.dmService.ListDMs(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, dms)
}
