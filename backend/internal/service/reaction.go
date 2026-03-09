package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository/postgres"
)

// ReactionBroadcaster - интерфейс для рассылки WS событий (избегаем import cycle)
type ReactionBroadcaster interface {
	BroadcastToWorkspace(workspaceID, event string, data interface{}, excludeUserID string)
}

type ReactionService struct {
	repo      *postgres.ReactionRepository
	wsHub     ReactionBroadcaster
}

func NewReactionService(repo *postgres.ReactionRepository, wsHub ReactionBroadcaster) *ReactionService {
	return &ReactionService{repo: repo, wsHub: wsHub}
}

func (s *ReactionService) Add(ctx context.Context, messageID, userID, emoji, workspaceID string) error {
	reaction := &domain.Reaction{
		ID:        uuid.New().String(),
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Add(ctx, reaction); err != nil {
		return err
	}

	s.wsHub.BroadcastToWorkspace(workspaceID, "reaction_added", map[string]interface{}{
		"message_id": messageID,
		"user_id":    userID,
		"emoji":      emoji,
	}, "")

	return nil
}

func (s *ReactionService) Remove(ctx context.Context, messageID, userID, emoji, workspaceID string) error {
	if err := s.repo.Remove(ctx, messageID, userID, emoji); err != nil {
		return err
	}

	s.wsHub.BroadcastToWorkspace(workspaceID, "reaction_removed", map[string]interface{}{
		"message_id": messageID,
		"user_id":    userID,
		"emoji":      emoji,
	}, "")

	return nil
}

func (s *ReactionService) GetGrouped(ctx context.Context, messageID, currentUserID string) ([]*domain.ReactionGroup, error) {
	return s.repo.GetGrouped(ctx, messageID, currentUserID)
}
