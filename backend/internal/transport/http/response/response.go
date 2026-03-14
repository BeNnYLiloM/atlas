package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/service"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success отправляет успешный ответ
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{Data: data})
}

// Created отправляет ответ о создании ресурса
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{Data: data})
}

// NoContent отправляет пустой ответ
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error отправляет ошибку с соответствующим статусом
func Error(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, service.ErrUserNotFound):
		status = http.StatusNotFound
		message = "user not found"
	case errors.Is(err, service.ErrUserAlreadyExists):
		status = http.StatusConflict
		message = "user already exists"
	case errors.Is(err, service.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		message = "invalid credentials"
	case errors.Is(err, service.ErrInvalidProfile):
		status = http.StatusBadRequest
		message = "invalid profile data"
	case errors.Is(err, service.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, service.ErrForbidden):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, service.ErrWorkspaceNotFound):
		status = http.StatusNotFound
		message = "workspace not found"
	case errors.Is(err, service.ErrChannelNotFound):
		status = http.StatusNotFound
		message = "channel not found"
	case errors.Is(err, service.ErrMessageNotFound):
		status = http.StatusNotFound
		message = "message not found"
	case errors.Is(err, service.ErrTaskNotFound):
		status = http.StatusNotFound
		message = "task not found"
	case errors.Is(err, service.ErrNotMember):
		status = http.StatusForbidden
		message = "not a member of workspace"
	case errors.Is(err, service.ErrProjectNotFound):
		status = http.StatusNotFound
		message = "project not found"
	case errors.Is(err, service.ErrNotProjectMember):
		status = http.StatusForbidden
		message = "not a member of this project"
	case errors.Is(err, service.ErrProjectArchived):
		status = http.StatusForbidden
		message = "project is archived"
	case errors.Is(err, service.ErrLastLead):
		status = http.StatusConflict
		message = "cannot remove the last lead from a project"
	}

	c.JSON(status, ErrorResponse{Error: message})
}

// BadRequest отправляет ошибку валидации
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request", Message: message})
}
