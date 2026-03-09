package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// Upload POST /api/v1/files/upload
func (h *FileHandler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploaded, err := h.fileService.Upload(c.Request.Context(), userID, header.Filename, file, header.Size, contentType)
	if err != nil {
		if err == service.ErrForbidden {
			response.Error(c, service.ErrForbidden)
			return
		}
		response.BadRequest(c, fmt.Sprintf("upload failed: %v", err))
		return
	}

	response.Success(c, uploaded)
}

// GetByID GET /api/v1/files/:id
func (h *FileHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	file, err := h.fileService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	response.Success(c, file)
}

// Delete DELETE /api/v1/files/:id
func (h *FileHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	if err := h.fileService.Delete(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
