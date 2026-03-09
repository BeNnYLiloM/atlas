package ws

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/your-org/atlas/backend/internal/service"
)

const (
	wsAuthProtocol        = "atlas.v1"
	wsTokenProtocolPrefix = "bearer."
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{wsAuthProtocol},
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

func extractTokenFromWebSocketRequest(r *http.Request) (string, error) {
	protocols := websocket.Subprotocols(r)
	if len(protocols) == 0 {
		return "", errors.New("missing websocket auth protocols")
	}

	hasAuthProtocol := false
	for _, protocol := range protocols {
		if protocol == wsAuthProtocol {
			hasAuthProtocol = true
			continue
		}
		if strings.HasPrefix(protocol, wsTokenProtocolPrefix) {
			if !hasAuthProtocol {
				return "", errors.New("missing websocket auth protocol")
			}
			token := strings.TrimPrefix(protocol, wsTokenProtocolPrefix)
			if token == "" {
				return "", errors.New("empty websocket token")
			}
			return token, nil
		}
	}

	if !hasAuthProtocol {
		return "", errors.New("missing websocket auth protocol")
	}

	return "", errors.New("missing websocket token protocol")
}

// HandleWebSocket обрабатывает WebSocket подключения
func (h *Handler) HandleWebSocket(c *gin.Context) {
	token, err := extractTokenFromWebSocketRequest(c.Request)
	if err != nil {
		log.Printf("[WS] Auth handshake failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing websocket authentication"})
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
