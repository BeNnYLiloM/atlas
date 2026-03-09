package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

type ReactionHandler struct {
	reactionService *service.ReactionService
}

func NewReactionHandler(reactionService *service.ReactionService) *ReactionHandler {
	return &ReactionHandler{reactionService: reactionService}
}

// Add POST /api/v1/messages/:id/reactions
func (h *ReactionHandler) Add(c *gin.Context) {
	userID := middleware.GetUserID(c)
	messageID := c.Param("id")
	workspaceID := c.Query("workspace_id")

	var body struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "emoji is required")
		return
	}

	if err := h.reactionService.Add(c.Request.Context(), messageID, userID, body.Emoji, workspaceID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"added": true})
}

// Remove DELETE /api/v1/messages/:id/reactions/:emoji
func (h *ReactionHandler) Remove(c *gin.Context) {
	userID := middleware.GetUserID(c)
	messageID := c.Param("id")
	emoji := c.Param("emoji")
	workspaceID := c.Query("workspace_id")

	if err := h.reactionService.Remove(c.Request.Context(), messageID, userID, emoji, workspaceID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"removed": true})
}

// GetReactions GET /api/v1/messages/:id/reactions
func (h *ReactionHandler) GetReactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	messageID := c.Param("id")

	groups, err := h.reactionService.GetGrouped(c.Request.Context(), messageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reactions"})
		return
	}

	response.Success(c, groups)
}
