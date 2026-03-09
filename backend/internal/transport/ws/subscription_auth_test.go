package ws

import (
	"context"
	"testing"
	"time"
)

type fakeAccessChecker struct {
	workspaceAllowed map[string]bool
	channelAllowed   map[string]bool
}

func (f *fakeAccessChecker) CanAccessWorkspace(ctx context.Context, workspaceID, userID string) (bool, error) {
	return f.workspaceAllowed[workspaceID], nil
}

func (f *fakeAccessChecker) CanAccessChannel(ctx context.Context, channelID, userID string) (bool, error) {
	return f.channelAllowed[channelID], nil
}

func TestClientHandleMessage_DeniesWorkspaceSubscriptionWithoutAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		ID:     "client-ws-deny",
		UserID: "user-1",
		hub:    hub,
		access: &fakeAccessChecker{workspaceAllowed: map[string]bool{"workspace-1": false}},
		send:   make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	client.handleMessage([]byte(`{"event":"subscribe_workspace","data":{"workspace_id":"workspace-1"}}`))
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.workspaces["workspace-1"][client.ID]
	hub.mu.RUnlock()

	if exists {
		t.Fatal("client should not be subscribed to workspace without access")
	}
}

func TestClientHandleMessage_AllowsWorkspaceSubscriptionWithAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		ID:     "client-ws-allow",
		UserID: "user-1",
		hub:    hub,
		access: &fakeAccessChecker{workspaceAllowed: map[string]bool{"workspace-1": true}},
		send:   make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	client.handleMessage([]byte(`{"event":"subscribe_workspace","data":{"workspace_id":"workspace-1"}}`))
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.workspaces["workspace-1"][client.ID]
	hub.mu.RUnlock()

	if !exists {
		t.Fatal("client should be subscribed to workspace when access is allowed")
	}
}

func TestClientHandleMessage_DeniesChannelSubscriptionWithoutAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		ID:     "client-ch-deny",
		UserID: "user-1",
		hub:    hub,
		access: &fakeAccessChecker{channelAllowed: map[string]bool{"channel-1": false}},
		send:   make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	client.handleMessage([]byte(`{"event":"subscribe","data":{"channel_id":"channel-1"}}`))
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.channels["channel-1"][client.ID]
	hub.mu.RUnlock()

	if exists {
		t.Fatal("client should not be subscribed to channel without access")
	}
}

func TestClientHandleMessage_AllowsChannelSubscriptionWithAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		ID:     "client-ch-allow",
		UserID: "user-1",
		hub:    hub,
		access: &fakeAccessChecker{channelAllowed: map[string]bool{"channel-1": true}},
		send:   make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	client.handleMessage([]byte(`{"event":"subscribe","data":{"channel_id":"channel-1"}}`))
	time.Sleep(20 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.channels["channel-1"][client.ID]
	hub.mu.RUnlock()

	if !exists {
		t.Fatal("client should be subscribed to channel when access is allowed")
	}
}
