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
