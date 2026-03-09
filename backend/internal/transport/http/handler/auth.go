package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.UserCreate true "Данные пользователя"
// @Success 201 {object} response.SuccessResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input domain.UserCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, tokens, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Login godoc
// @Summary Авторизация пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.UserLogin true "Данные для входа"
// @Success 200 {object} response.SuccessResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input domain.UserLogin
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, tokens, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Me godoc
// @Summary Получить текущего пользователя
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// SearchByEmail ищет пользователя по email
// @Summary Поиск пользователя по email
// @Tags users
// @Security Bearer
// @Produce json
// @Param email query string true "Email пользователя"
// @Success 200 {object} response.SuccessResponse
// @Router /users/search [get]
func (h *AuthHandler) SearchByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "email is required")
		return
	}

	user, err := h.authService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// RegisterRoutes регистрирует маршруты
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/me", authMiddleware, h.Me)
	}

	users := r.Group("/users", authMiddleware)
	{
		users.GET("/search", h.SearchByEmail)
	}
}

