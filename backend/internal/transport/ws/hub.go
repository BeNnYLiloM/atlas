package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// Hub управляет WebSocket соединениями
type Hub struct {
	// Клиенты по workspace: workspaceID -> map[clientID]Client (основное)
	workspaces map[string]map[string]*Client

	// Клиенты по каналам: channelID -> map[clientID]Client (для typing indicators)
	channels map[string]map[string]*Client

	// Клиенты по пользователям: userID -> []*Client (один юзер может иметь несколько соединений)
	users map[string][]*Client

	// Регистрация нового клиента
	register chan *Client

	// Отключение клиента
	unregister chan *Client

	// Подписка на workspace/channel
	subscribe chan *Subscription

	// Отписка от workspace/channel
	unsubscribe chan *Subscription

	// Сообщение для рассылки
	broadcast chan *BroadcastMessage

	mu sync.RWMutex
}

type Subscription struct {
	Client      *Client
	WorkspaceID string // Для workspace подписок
	ChannelID   string // Для channel подписок (typing)
}

type BroadcastMessage struct {
	WorkspaceID   string      // Для workspace broadcast
	ChannelID     string      // Для channel broadcast (опционально)
	Event         string
	Data          interface{}
	ExcludeUserID string // UserID клиента, которого исключить из рассылки
}

// WSMessage - формат WebSocket сообщения (входящее)
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// OutgoingMessage - исходящее сообщение (совместимо с frontend)
type OutgoingMessage struct {
	Type    string      `json:"type"`    // Используем "type" вместо "event" для совместимости с frontend
	Payload interface{} `json:"payload"` // Используем "payload" вместо "data"
}

func NewHub() *Hub {
	return &Hub{
		workspaces:  make(map[string]map[string]*Client),
		channels:    make(map[string]map[string]*Client),
		users:       make(map[string][]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		subscribe:   make(chan *Subscription),
		unsubscribe: make(chan *Subscription),
		broadcast:   make(chan *BroadcastMessage, 256),
	}
}

// Run запускает главный цикл обработки событий
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case sub := <-h.subscribe:
			h.handleSubscribe(sub)

		case sub := <-h.unsubscribe:
			h.handleUnsubscribe(sub)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	h.users[client.UserID] = append(h.users[client.UserID], client)
	h.mu.Unlock()

	log.Printf("Client registered: userID=%s, clientID=%s", client.UserID, client.ID)
}

// BroadcastPresence рассылает присутствие пользователя всем workspace, на которые он подписан
func (h *Hub) BroadcastPresence(userID, status string) {
	h.mu.RLock()
	// Собираем все workspace, в которых есть этот пользователь
	affectedWorkspaces := make(map[string]bool)
	for workspaceID, clients := range h.workspaces {
		for _, c := range clients {
			if c.UserID == userID {
				affectedWorkspaces[workspaceID] = true
				break
			}
		}
	}
	h.mu.RUnlock()

	for workspaceID := range affectedWorkspaces {
		h.BroadcastToWorkspace(workspaceID, "presence", map[string]interface{}{
			"user_id": userID,
			"status":  status,
		}, "")
	}
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()

	// Удаляем из всех workspace
	for workspaceID, clients := range h.workspaces {
		if _, ok := clients[client.ID]; ok {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(h.workspaces, workspaceID)
			}
		}
	}

	// Удаляем из всех каналов
	for channelID, clients := range h.channels {
		if _, ok := clients[client.ID]; ok {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(h.channels, channelID)
			}
		}
	}

	// Удаляем из списка пользователей
	var remaining int
	if userClients, ok := h.users[client.UserID]; ok {
		for i, c := range userClients {
			if c.ID == client.ID {
				h.users[client.UserID] = append(userClients[:i], userClients[i+1:]...)
				break
			}
		}
		if len(h.users[client.UserID]) == 0 {
			delete(h.users, client.UserID)
		}
		remaining = len(h.users[client.UserID])
	}

	h.mu.Unlock()

	close(client.send)
	log.Printf("Client unregistered: userID=%s, clientID=%s", client.UserID, client.ID)

	if remaining == 0 {
		go h.BroadcastPresence(client.UserID, "offline")
	}
}

func (h *Hub) handleSubscribe(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Подписка на workspace (основное)
	if sub.WorkspaceID != "" {
		if h.workspaces[sub.WorkspaceID] == nil {
			h.workspaces[sub.WorkspaceID] = make(map[string]*Client)
		}

		// Собираем уже онлайн-пользователей до добавления нового клиента
		onlineUserIDs := make(map[string]bool)
		for _, c := range h.workspaces[sub.WorkspaceID] {
			onlineUserIDs[c.UserID] = true
		}

		h.workspaces[sub.WorkspaceID][sub.Client.ID] = sub.Client
		log.Printf("[Hub] Client subscribed to workspace: userID=%s, workspaceID=%s", sub.Client.UserID, sub.WorkspaceID)

		// Отправляем новому клиенту текущие статусы всех уже онлайн-участников
		newClient := sub.Client
		go func(onlineIDs map[string]bool) {
			outMsg := OutgoingMessage{}
			for uid := range onlineIDs {
				outMsg.Type = "presence"
				outMsg.Payload = map[string]interface{}{
					"user_id": uid,
					"status":  "online",
				}
				data, err := json.Marshal(outMsg)
				if err != nil {
					continue
				}
				select {
				case newClient.send <- data:
				default:
				}
			}
		}(onlineUserIDs)

		// Broadcast presence online всем остальным в воркспейсе
		go func(workspaceID, userID string) {
			h.BroadcastToWorkspace(workspaceID, "presence", map[string]interface{}{
				"user_id": userID,
				"status":  "online",
			}, "")
		}(sub.WorkspaceID, sub.Client.UserID)
	}

	// Подписка на channel (для typing indicators)
	if sub.ChannelID != "" {
		if h.channels[sub.ChannelID] == nil {
			h.channels[sub.ChannelID] = make(map[string]*Client)
		}
		h.channels[sub.ChannelID][sub.Client.ID] = sub.Client
		log.Printf("[Hub] Client subscribed to channel: userID=%s, channelID=%s", sub.Client.UserID, sub.ChannelID)
	}
}

func (h *Hub) handleUnsubscribe(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Отписка от workspace
	if sub.WorkspaceID != "" {
		if clients, ok := h.workspaces[sub.WorkspaceID]; ok {
			delete(clients, sub.Client.ID)
			if len(clients) == 0 {
				delete(h.workspaces, sub.WorkspaceID)
			}
			log.Printf("[Hub] Client unsubscribed from workspace: userID=%s, workspaceID=%s", sub.Client.UserID, sub.WorkspaceID)
		}
	}

	// Отписка от channel
	if sub.ChannelID != "" {
		if clients, ok := h.channels[sub.ChannelID]; ok {
			delete(clients, sub.Client.ID)
			if len(clients) == 0 {
				delete(h.channels, sub.ChannelID)
			}
			log.Printf("[Hub] Client unsubscribed from channel: userID=%s, channelID=%s", sub.Client.UserID, sub.ChannelID)
		}
	}
}

func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	
	var clients map[string]*Client
	var target string
	
	// Определяем куда broadcast: workspace или channel
	if msg.WorkspaceID != "" {
		clients = h.workspaces[msg.WorkspaceID]
		target = fmt.Sprintf("workspace %s", msg.WorkspaceID)
	} else if msg.ChannelID != "" {
		clients = h.channels[msg.ChannelID]
		target = fmt.Sprintf("channel %s", msg.ChannelID)
	}
	
	h.mu.RUnlock()

	if clients == nil || len(clients) == 0 {
		log.Printf("[Hub] No subscribers for %s", target)
		return
	}

	log.Printf("[Hub] Broadcasting event '%s' to %s (%d subscribers)", msg.Event, target, len(clients))

	outMsg := OutgoingMessage{
		Type:    msg.Event,
		Payload: msg.Data,
	}

	data, err := json.Marshal(outMsg)
	if err != nil {
		log.Printf("[Hub] Error marshaling broadcast message: %v", err)
		return
	}

	log.Printf("[Hub] Broadcast data: %s", string(data))

	sentCount := 0
	for clientID, client := range clients {
		// Исключаем по UserID, а не по ClientID
		if msg.ExcludeUserID != "" && client.UserID == msg.ExcludeUserID {
			log.Printf("[Hub] Skipping excluded user: %s (clientID: %s)", client.UserID, clientID)
			continue
		}
		select {
		case client.send <- data:
			sentCount++
			log.Printf("[Hub] Sent to client %s (userID: %s)", clientID, client.UserID)
		default:
			// Буфер заполнен, пропускаем
			log.Printf("[Hub] Client buffer full, skipping: %s", clientID)
		}
	}
	
	log.Printf("[Hub] Broadcast complete: sent to %d/%d clients", sentCount, len(clients))
}

// Register регистрирует клиента
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister отключает клиента
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SubscribeToWorkspace подписывает клиента на workspace
func (h *Hub) SubscribeToWorkspace(client *Client, workspaceID string) {
	h.subscribe <- &Subscription{Client: client, WorkspaceID: workspaceID}
}

// UnsubscribeFromWorkspace отписывает клиента от workspace
func (h *Hub) UnsubscribeFromWorkspace(client *Client, workspaceID string) {
	h.unsubscribe <- &Subscription{Client: client, WorkspaceID: workspaceID}
}

// Subscribe подписывает клиента на канал (для typing)
func (h *Hub) Subscribe(client *Client, channelID string) {
	h.subscribe <- &Subscription{Client: client, ChannelID: channelID}
}

// Unsubscribe отписывает клиента от канала
func (h *Hub) Unsubscribe(client *Client, channelID string) {
	h.unsubscribe <- &Subscription{Client: client, ChannelID: channelID}
}

// BroadcastToWorkspace рассылает сообщение всем подписчикам workspace
func (h *Hub) BroadcastToWorkspace(workspaceID, event string, data interface{}, excludeUserID string) {
	h.broadcast <- &BroadcastMessage{
		WorkspaceID:   workspaceID,
		Event:         event,
		Data:          data,
		ExcludeUserID: excludeUserID,
	}
}

// BroadcastToUsers рассылает сообщение конкретному списку пользователей (если они онлайн)
func (h *Hub) BroadcastToUsers(userIDs []string, event string, data interface{}) {
	outMsg := OutgoingMessage{
		Type:    event,
		Payload: data,
	}
	msgData, err := json.Marshal(outMsg)
	if err != nil {
		log.Printf("[Hub] BroadcastToUsers marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range userIDs {
		for _, client := range h.users[uid] {
			select {
			case client.send <- msgData:
			default:
			}
		}
	}
}

// Broadcast рассылает сообщение всем подписчикам канала
func (h *Hub) Broadcast(channelID, event string, data interface{}, excludeUserID string) {
	h.broadcast <- &BroadcastMessage{
		ChannelID:     channelID,
		Event:         event,
		Data:          data,
		ExcludeUserID: excludeUserID,
	}
}

// SendToUser отправляет сообщение конкретному пользователю
func (h *Hub) SendToUser(userID, event string, data interface{}) {
	h.mu.RLock()
	clients := h.users[userID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	outMsg := OutgoingMessage{
		Type:    event,
		Payload: data,
	}

	msgData, err := json.Marshal(outMsg)
	if err != nil {
		return
	}

	for _, client := range clients {
		select {
		case client.send <- msgData:
		default:
		}
	}
}

