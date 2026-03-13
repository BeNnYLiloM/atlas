package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/config"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

const refreshCookiePath = "/api/v1/auth"

type AuthHandler struct {
	authService *service.AuthService
	fileService *service.FileService
	jwtConfig   config.JWTConfig
	wsHub       *ws.Hub
}

func NewAuthHandler(authService *service.AuthService, fileService *service.FileService, jwtConfig config.JWTConfig, wsHub *ws.Hub) *AuthHandler {
	return &AuthHandler{authService: authService, fileService: fileService, jwtConfig: jwtConfig, wsHub: wsHub}
}

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

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(h.jwtConfig.RefreshCookieName)
	if err := h.authService.Logout(c.Request.Context(), refreshToken); err != nil {
		response.Error(c, err)
		return
	}

	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input domain.UserUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if input.DisplayName != nil && strings.TrimSpace(*input.DisplayName) == "" {
		response.BadRequest(c, "display_name must not be empty")
		return
	}

	updatedUser, err := h.authService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, updatedUser)
}

const maxAvatarBytes = 10 << 20 // 10 МБ

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	if h.fileService == nil {
		response.BadRequest(c, "file storage unavailable")
		return
	}

	if err := c.Request.ParseMultipartForm(maxAvatarBytes); err != nil {
		response.BadRequest(c, "avatar file is required")
		return
	}

	userID := middleware.GetUserID(c)
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "avatar file is required")
		return
	}
	defer file.Close()

	if header.Size > maxAvatarBytes {
		response.BadRequest(c, "avatar must not exceed 10 MB")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		response.BadRequest(c, "avatar must be an image")
		return
	}

	uploaded, err := h.fileService.Upload(c.Request.Context(), userID, header.Filename, file, header.Size, contentType)
	if err != nil {
		response.BadRequest(c, fmt.Sprintf("upload failed: %v", err))
		return
	}

	updatedUser, err := h.authService.UpdateProfile(c.Request.Context(), userID, domain.UserUpdate{
		AvatarURL: &uploaded.URL,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, updatedUser)
}

type sessionResponse struct {
	*domain.AuthSession
	IsCurrent bool `json:"is_current"`
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	currentSessionID := middleware.GetSessionID(c)

	sessions, err := h.authService.ListActiveSessions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]sessionResponse, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, sessionResponse{
			AuthSession: s,
			IsCurrent:   s.ID == currentSessionID,
		})
	}
	response.Success(c, result)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input domain.UserChangePassword
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, input); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

type updateStatusRequest struct {
	Status       domain.UserStatus `json:"status"`
	CustomStatus *string           `json:"custom_status"`
}

func (h *AuthHandler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input updateStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.UpdateStatus(c.Request.Context(), userID, input.Status, input.CustomStatus)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Рассылаем presence-событие всем участникам workspace'ов этого пользователя
	if h.wsHub != nil {
		go h.wsHub.BroadcastPresence(userID, string(input.Status))
	}

	response.Success(c, user)
}

type deleteAccountRequest struct {
	Password string `json:"password"`
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input deleteAccountRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if input.Password == "" {
		response.BadRequest(c, "password is required")
		return
	}

	if err := h.authService.DeleteAccount(c.Request.Context(), userID, input.Password); err != nil {
		response.Error(c, err)
		return
	}

	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	currentSessionID := middleware.GetSessionID(c)
	sessionID := c.Param("id")

	if sessionID == currentSessionID {
		response.BadRequest(c, "cannot revoke current session; use logout instead")
		return
	}

	revoked, err := h.authService.RevokeSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if !revoked {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	response.NoContent(c)
}

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
		auth.PATCH("/me", authMiddleware, h.UpdateMe)
		auth.POST("/me/avatar", authMiddleware, h.UploadAvatar)
		auth.PATCH("/me/password", authMiddleware, h.ChangePassword)
		auth.PATCH("/me/status", authMiddleware, h.UpdateStatus)
		auth.DELETE("/me", authMiddleware, h.DeleteAccount)
		auth.GET("/me/sessions", authMiddleware, h.ListSessions)
		auth.DELETE("/me/sessions/:id", authMiddleware, h.RevokeSession)
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
