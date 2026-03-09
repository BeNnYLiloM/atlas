package service

import (
	"context"
	"fmt"
	"time"

	lkauth "github.com/livekit/protocol/auth"
	"github.com/your-org/atlas/backend/internal/config"
)

type LiveKitService struct {
	cfg config.LiveKitConfig
}

func NewLiveKitService(cfg config.LiveKitConfig) *LiveKitService {
	return &LiveKitService{cfg: cfg}
}

type CallToken struct {
	Token    string `json:"token"`
	RoomName string `json:"room_name"`
	URL      string `json:"url"`
}

// CreateToken создаёт JWT токен для подключения к комнате LiveKit
func (s *LiveKitService) CreateToken(ctx context.Context, roomName, userID, displayName string, canPublish bool) (*CallToken, error) {
	at := lkauth.NewAccessToken(s.cfg.APIKey, s.cfg.APISecret)

	grant := &lkauth.VideoGrant{
		RoomJoin:     true,
		Room:         roomName,
		CanPublish:   &canPublish,
		CanSubscribe: func() *bool { t := true; return &t }(),
	}

	at.AddGrant(grant).
		SetIdentity(userID).
		SetName(displayName).
		SetValidFor(2 * time.Hour)

	token, err := at.ToJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate livekit token: %w", err)
	}

	return &CallToken{
		Token:    token,
		RoomName: roomName,
		URL:      s.cfg.URL,
	}, nil
}

// CreateRoomName возвращает стабильное имя комнаты для канала (как в Discord)
func (s *LiveKitService) CreateRoomName(channelID string) string {
	return fmt.Sprintf("channel-%s", channelID)
}
