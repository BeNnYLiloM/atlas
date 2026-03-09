package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/config"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
)

const refreshCookiePath = "/api/v1/auth"

type AuthHandler struct {
	authService *service.AuthService
	jwtConfig   config.JWTConfig
}

func NewAuthHandler(authService *service.AuthService, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{authService: authService, jwtConfig: jwtConfig}
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

	user, tokens, refreshToken, err := h.authService.Register(c.Request.Context(), input, sessionMetadataFromRequest(c))
	if err != nil {
		response.Error(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
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

	user, tokens, refreshToken, err := h.authService.Login(c.Request.Context(), input, sessionMetadataFromRequest(c))
	if err != nil {
		response.Error(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
	response.Success(c, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Refresh godoc
// @Summary Обновить access token по refresh cookie
// @Tags auth
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, _ := c.Cookie(h.jwtConfig.RefreshCookieName)
	tokens, nextRefreshToken, err := h.authService.Refresh(c.Request.Context(), refreshToken, sessionMetadataFromRequest(c))
	if err != nil {
		h.clearRefreshCookie(c)
		response.Error(c, err)
		return
	}

	h.setRefreshCookie(c, nextRefreshToken)
	response.Success(c, gin.H{
		"tokens": tokens,
	})
}

// Logout godoc
// @Summary Завершить текущую сессию
// @Tags auth
// @Produce json
// @Success 204
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(h.jwtConfig.RefreshCookieName)
	if err := h.authService.Logout(c.Request.Context(), refreshToken); err != nil {
		response.Error(c, err)
		return
	}

	h.clearRefreshCookie(c)
	response.NoContent(c)
}

// LogoutAll godoc
// @Summary Завершить все активные сессии
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 204
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	h.clearRefreshCookie(c)
	response.NoContent(c)
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
	loginLimiter := middleware.NewRateLimiter(10, time.Minute)
	registerLimiter := middleware.NewRateLimiter(5, time.Minute)
	refreshLimiter := middleware.NewRateLimiter(30, time.Minute)

	auth := r.Group("/auth")
	{
		auth.POST("/register", registerLimiter, h.Register)
		auth.POST("/login", loginLimiter, h.Login)
		auth.POST("/refresh", refreshLimiter, h.Refresh)
		auth.POST("/logout", h.Logout)
		auth.POST("/logout-all", authMiddleware, h.LogoutAll)
		auth.GET("/me", authMiddleware, h.Me)
	}

	users := r.Group("/users", authMiddleware)
	{
		users.GET("/search", h.SearchByEmail)
	}
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		h.jwtConfig.RefreshCookieName,
		token,
		h.authService.RefreshCookieMaxAgeSeconds(),
		refreshCookiePath,
		h.jwtConfig.RefreshCookieDomain,
		h.jwtConfig.RefreshCookieSecure,
		true,
	)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		h.jwtConfig.RefreshCookieName,
		"",
		-1,
		refreshCookiePath,
		h.jwtConfig.RefreshCookieDomain,
		h.jwtConfig.RefreshCookieSecure,
		true,
	)
}

func sessionMetadataFromRequest(c *gin.Context) service.AuthSessionMetadata {
	return service.AuthSessionMetadata{
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	}
}
