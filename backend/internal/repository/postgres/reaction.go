package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type ReactionRepository struct {
	db *pgxpool.Pool
}

func NewReactionRepository(db *pgxpool.Pool) *ReactionRepository {
	return &ReactionRepository{db: db}
}

func (r *ReactionRepository) Add(ctx context.Context, reaction *domain.Reaction) error {
	query := `
		INSERT INTO message_reactions (id, message_id, user_id, emoji, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query,
		reaction.ID, reaction.MessageID, reaction.UserID, reaction.Emoji, reaction.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("ReactionRepository.Add: %w", err)
	}
	return nil
}

func (r *ReactionRepository) Remove(ctx context.Context, messageID, userID, emoji string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM message_reactions WHERE message_id=$1 AND user_id=$2 AND emoji=$3`,
		messageID, userID, emoji,
	)
	if err != nil {
		return fmt.Errorf("ReactionRepository.Remove: %w", err)
	}
	return nil
}

// GetGrouped возвращает реакции сгруппированные по emoji
func (r *ReactionRepository) GetGrouped(ctx context.Context, messageID, currentUserID string) ([]*domain.ReactionGroup, error) {
	query := `
		SELECT emoji, array_agg(user_id) as user_ids, COUNT(*) as count
		FROM message_reactions
		WHERE message_id = $1
		GROUP BY emoji
		ORDER BY MIN(created_at) ASC
	`
	rows, err := r.db.Query(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("ReactionRepository.GetGrouped: %w", err)
	}
	defer rows.Close()

	var groups []*domain.ReactionGroup
	for rows.Next() {
		g := &domain.ReactionGroup{}
		if err := rows.Scan(&g.Emoji, &g.UserIDs, &g.Count); err != nil {
			return nil, fmt.Errorf("ReactionRepository.GetGrouped scan: %w", err)
		}
		for _, uid := range g.UserIDs {
			if uid == currentUserID {
				g.Mine = true
				break
			}
		}
		groups = append(groups, g)
	}
	return groups, nil
}
