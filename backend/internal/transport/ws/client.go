package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Время ожидания записи сообщения
	writeWait = 10 * time.Second

	// Время ожидания pong от клиента
	pongWait = 60 * time.Second

	// Интервал отправки ping
	pingPeriod = (pongWait * 9) / 10

	// Максимальный размер сообщения
	maxMessageSize = 4096
)

// Client представляет WebSocket клиента
type Client struct {
	ID          string
	UserID      string
	WorkspaceID string // текущий workspace для routing typing events
	hub         *Hub
	access      AccessChecker
	conn        *websocket.Conn
	send        chan []byte
}

// NewClient создает нового клиента
func NewClient(hub *Hub, conn *websocket.Conn, userID string, access AccessChecker) *Client {
	return &Client{
		ID:     uuid.New().String(),
		UserID: userID,
		hub:    hub,
		access: access,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

// ReadPump читает сообщения от клиента
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump отправляет сообщения клиенту
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Отправляем все накопившиеся сообщения
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage обрабатывает входящее сообщение
func (c *Client) handleMessage(message []byte) {
	log.Printf("[WS Client %s] Received raw message: %s", c.ID, string(message))

	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("[WS Client %s] Error unmarshaling message: %v", c.ID, err)
		return
	}

	log.Printf("[WS Client %s] Parsed event: %s, data: %s", c.ID, msg.Event, string(msg.Data))

	switch msg.Event {
	case "subscribe_workspace":
		var data struct {
			WorkspaceID string `json:"workspace_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[WS Client %s] Error parsing subscribe_workspace data: %v", c.ID, err)
			return
		}
		log.Printf("[WS Client %s] Subscribing to workspace: %s", c.ID, data.WorkspaceID)
		allowed, err := c.access.CanAccessWorkspace(context.Background(), data.WorkspaceID, c.UserID)
		if err != nil {
			log.Printf("[WS Client %s] Workspace access check failed: %v", c.ID, err)
			return
		}
		if !allowed {
			log.Printf("[WS Client %s] Workspace access denied: %s", c.ID, data.WorkspaceID)
			return
		}
		c.WorkspaceID = data.WorkspaceID
		c.hub.SubscribeToWorkspace(c, data.WorkspaceID)

	case "unsubscribe_workspace":
		var data struct {
			WorkspaceID string `json:"workspace_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[WS Client %s] Error parsing unsubscribe_workspace data: %v", c.ID, err)
			return
		}
		log.Printf("[WS Client %s] Unsubscribing from workspace: %s", c.ID, data.WorkspaceID)
		c.hub.UnsubscribeFromWorkspace(c, data.WorkspaceID)

	case "subscribe":
		var data struct {
			ChannelID string `json:"channel_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[WS Client %s] Error parsing subscribe data: %v", c.ID, err)
			return
		}
		log.Printf("[WS Client %s] Subscribing to channel: %s", c.ID, data.ChannelID)
		allowed, err := c.access.CanAccessChannel(context.Background(), data.ChannelID, c.UserID)
		if err != nil {
			log.Printf("[WS Client %s] Channel access check failed: %v", c.ID, err)
			return
		}
		if !allowed {
			log.Printf("[WS Client %s] Channel access denied: %s", c.ID, data.ChannelID)
			return
		}
		c.hub.Subscribe(c, data.ChannelID)

	case "unsubscribe":
		var data struct {
			ChannelID string `json:"channel_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[WS Client %s] Error parsing unsubscribe data: %v", c.ID, err)
			return
		}
		log.Printf("[WS Client %s] Unsubscribing from channel: %s", c.ID, data.ChannelID)
		c.hub.Unsubscribe(c, data.ChannelID)

	case "typing":
		var data struct {
			ChannelID string `json:"channel_id"`
			Typing    bool   `json:"typing"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("[WS Client %s] Error parsing typing data: %v", c.ID, err)
			return
		}
		log.Printf("[WS Client %s] User typing event: channelID=%s, workspaceID=%s, typing=%v", c.ID, data.ChannelID, c.WorkspaceID, data.Typing)
		if c.WorkspaceID == "" {
			log.Printf("[WS Client %s] No workspaceID for typing event, dropping", c.ID)
			return
		}
		allowed, err := c.access.CanAccessChannel(context.Background(), data.ChannelID, c.UserID)
		if err != nil {
			log.Printf("[WS Client %s] Typing access check failed: %v", c.ID, err)
			return
		}
		if !allowed {
			log.Printf("[WS Client %s] Typing access denied for channel: %s", c.ID, data.ChannelID)
			return
		}
		c.hub.BroadcastToWorkspace(c.WorkspaceID, "typing", map[string]interface{}{
			"user_id":    c.UserID,
			"channel_id": data.ChannelID,
			"typing":     data.Typing,
		}, c.UserID)

	default:
		log.Printf("[WS Client %s] Unknown event: %s", c.ID, msg.Event)
	}
}
