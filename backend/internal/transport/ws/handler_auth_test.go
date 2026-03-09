package ws

import (
	"net/http"
	"testing"
)

func TestExtractTokenFromWebSocketRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Sec-WebSocket-Protocol", wsAuthProtocol+", "+wsTokenProtocolPrefix+"jwt-token")

	token, err := extractTokenFromWebSocketRequest(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if token != "jwt-token" {
		t.Fatalf("expected jwt-token, got %q", token)
	}
}

func TestExtractTokenFromWebSocketRequest_DeniesMissingToken(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Sec-WebSocket-Protocol", wsAuthProtocol)

	_, err = extractTokenFromWebSocketRequest(req)
	if err == nil {
		t.Fatal("expected error when token protocol is missing")
	}
}

func TestExtractTokenFromWebSocketRequest_DeniesMissingAuthProtocol(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/ws?token=legacy", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Sec-WebSocket-Protocol", wsTokenProtocolPrefix+"jwt-token")

	_, err = extractTokenFromWebSocketRequest(req)
	if err == nil {
		t.Fatal("expected error when auth protocol is missing")
	}
}

func TestHandlerCheckOrigin_AllowsConfiguredOrigin(t *testing.T) {
	handler := NewHandler(nil, nil, nil, []string{"https://app.atlas.local"})
	req, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Host = "api.atlas.local"
	req.Header.Set("Origin", "https://app.atlas.local")

	if !handler.checkOrigin(req) {
		t.Fatal("expected configured origin to be allowed")
	}
}

func TestHandlerCheckOrigin_DeniesUnknownOrigin(t *testing.T) {
	handler := NewHandler(nil, nil, nil, []string{"https://app.atlas.local"})
	req, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Host = "api.atlas.local"
	req.Header.Set("Origin", "https://evil.example")

	if handler.checkOrigin(req) {
		t.Fatal("expected unknown origin to be denied")
	}
}
