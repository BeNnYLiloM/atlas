package ws

import (
	"encoding/json"
	"testing"
	"time"
)

// TestHub_RegisterUnregister проверяет регистрацию и отключение клиентов
func TestHub_RegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		ID:     "client-1",
		UserID: "user-1",
		hub:    hub,
		send:   make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.users["user-1"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("client should be registered in users map")
	}

	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	remaining := len(hub.users["user-1"])
	hub.mu.RUnlock()

	if remaining != 0 {
		t.Errorf("expected 0 clients after unregister, got %d", remaining)
	}
}

// TestHub_WorkspaceSubscription проверяет подписку на workspace
func TestHub_WorkspaceSubscription(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		ID:     "client-1",
		UserID: "user-1",
		hub:    hub,
		send:   make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	hub.SubscribeToWorkspace(client, "workspace-1")
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	wsClients := hub.workspaces["workspace-1"]
	hub.mu.RUnlock()

	if len(wsClients) != 1 {
		t.Errorf("expected 1 subscriber for workspace, got %d", len(wsClients))
	}

	hub.UnsubscribeFromWorkspace(client, "workspace-1")
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	wsClients = hub.workspaces["workspace-1"]
	hub.mu.RUnlock()

	if len(wsClients) != 0 {
		t.Errorf("expected 0 subscribers after unsubscribe, got %d", len(wsClients))
	}
}

// drainUntilEvent дренирует канал до получения события указанного типа или таймаута
func drainUntilEvent(send chan []byte, wantType string, timeout time.Duration) bool {
	deadline := time.After(timeout)
	for {
		select {
		case msg := <-send:
			var event OutgoingMessage
			if err := json.Unmarshal(msg, &event); err != nil {
				continue
			}
			if event.Type == wantType {
				return true
			}
		case <-deadline:
			return false
		}
	}
}

// TestHub_BroadcastToWorkspace проверяет рассылку событий
func TestHub_BroadcastToWorkspace(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := &Client{ID: "c1b", UserID: "u1b", hub: hub, send: make(chan []byte, 256)}
	client2 := &Client{ID: "c2b", UserID: "u2b", hub: hub, send: make(chan []byte, 256)}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	hub.SubscribeToWorkspace(client1, "ws-broadcast")
	hub.SubscribeToWorkspace(client2, "ws-broadcast")
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToWorkspace("ws-broadcast", "test_event", map[string]string{"key": "value"}, "")
	time.Sleep(50 * time.Millisecond)

	// Оба клиента должны получить test_event (среди других сообщений)
	for i, c := range []*Client{client1, client2} {
		if !drainUntilEvent(c.send, "test_event", 300*time.Millisecond) {
			t.Errorf("client %d: did not receive 'test_event' broadcast", i+1)
		}
	}
}

// TestHub_BroadcastExcludesUser проверяет исключение пользователя из рассылки
func TestHub_BroadcastExcludesUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	sender := &Client{ID: "s-ex", UserID: "u-sender-ex", hub: hub, send: make(chan []byte, 256)}
	receiver := &Client{ID: "r-ex", UserID: "u-recv-ex", hub: hub, send: make(chan []byte, 256)}

	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	hub.SubscribeToWorkspace(sender, "ws-excl")
	hub.SubscribeToWorkspace(receiver, "ws-excl")
	time.Sleep(50 * time.Millisecond)

	// Дренируем presence события
	drainChannel(sender.send, 100*time.Millisecond)

	// Broadcast исключает отправителя
	hub.BroadcastToWorkspace("ws-excl", "exclusive_msg", "data", "u-sender-ex")
	time.Sleep(50 * time.Millisecond)

	// sender не должен получить exclusive_msg
	if hasEvent(sender.send, "exclusive_msg") {
		t.Error("sender should not receive its own broadcast")
	}

	// receiver должен получить
	if !drainUntilEvent(receiver.send, "exclusive_msg", 300*time.Millisecond) {
		t.Error("receiver did not receive broadcast message")
	}
}

func drainChannel(ch chan []byte, timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case <-ch:
		case <-deadline:
			return
		}
	}
}

func hasEvent(ch chan []byte, eventType string) bool {
	for {
		select {
		case msg := <-ch:
			var event OutgoingMessage
			if json.Unmarshal(msg, &event) == nil && event.Type == eventType {
				return true
			}
		default:
			return false
		}
	}
}
