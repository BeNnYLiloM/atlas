package ws

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/your-org/atlas/backend/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		originURL, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if strings.EqualFold(originURL.Host, r.Host) {
			return true
		}

		host := originURL.Hostname()
		return host == "localhost" || host == "127.0.0.1"
	},
}

type AccessChecker interface {
	CanAccessWorkspace(ctx context.Context, workspaceID, userID string) (bool, error)
	CanAccessChannel(ctx context.Context, channelID, userID string) (bool, error)
}

type Handler struct {
	hub         *Hub
	authService *service.AuthService
	access      AccessChecker
}

func NewHandler(hub *Hub, authService *service.AuthService, access AccessChecker) *Handler {
	return &Handler{
		hub:         hub,
		authService: authService,
		access:      access,
	}
}

// HandleWebSocket обрабатывает WebSocket подключения
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Получаем токен из query параметра
	token := c.Query("token")
	if token == "" {
		log.Println("[WS] Missing token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Валидируем токен
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		log.Printf("[WS] Invalid token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	log.Printf("[WS] User %s attempting to connect", claims.UserID)

	// Апгрейдим соединение до WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] WebSocket upgrade error: %v", err)
		return
	}

	client := NewClient(h.hub, conn, claims.UserID, h.access)
	h.hub.Register(client)

	log.Printf("[WS] Client connected: userID=%s, clientID=%s", claims.UserID, client.ID)

	// Запускаем горутины для чтения и записи
	go client.WritePump()
	go client.ReadPump()
}

// RegisterRoutes регистрирует WebSocket маршрут
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/ws", h.HandleWebSocket)
}
