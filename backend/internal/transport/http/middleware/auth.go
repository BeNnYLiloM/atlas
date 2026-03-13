package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/atlas/backend/internal/service"
)

const (
	AuthorizationHeader = "Authorization"
	UserIDKey           = "userID"
	UserEmailKey        = "userEmail"
	SessionIDKey        = "sessionID"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(AuthorizationHeader)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(SessionIDKey, claims.SessionID)
		c.Next()
	}
}

// GetUserID извлекает userID из контекста. Паникует только если middleware не был применён к маршруту.
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		panic("GetUserID called on route without AuthMiddleware")
	}
	id, ok := userID.(string)
	if !ok || id == "" {
		panic("GetUserID: userID in context is not a valid string")
	}
	return id
}

// GetSessionID извлекает sessionID из контекста JWT claims.
func GetSessionID(c *gin.Context) string {
	sessionID, exists := c.Get(SessionIDKey)
	if !exists {
		return ""
	}
	id, ok := sessionID.(string)
	if !ok {
		return ""
	}
	return id
}

