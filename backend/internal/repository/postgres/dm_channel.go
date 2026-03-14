package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type DMChannelRepo struct {
	db *pgxpool.Pool
}

func NewDMChannelRepo(db *pgxpool.Pool) *DMChannelRepo {
	return &DMChannelRepo{db: db}
}

// GetOrCreate атомарно возвращает существующий DM-канал или создаёт новый.
// Инвариант: user1_id < user2_id обеспечивает уникальность пары без дублей.
func (r *DMChannelRepo) GetOrCreate(ctx context.Context, workspaceID, userID1, userID2 string) (*domain.Channel, error) {
	// Нормализуем порядок — меньший UUID идёт первым
	lo, hi := userID1, userID2
	if lo > hi {
		lo, hi = hi, lo
	}

	// Сначала пробуем найти существующий
	existing, err := r.findExisting(ctx, workspaceID, lo, hi)
	if err != nil {
		return nil, fmt.Errorf("dm: find existing: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Создаём в транзакции: channel + dm_channels + channel_members (оба участника)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("dm: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	channelID := uuid.New().String()
	ch := &domain.Channel{}

	err = tx.QueryRow(ctx, `
		INSERT INTO channels (id, workspace_id, name, type, is_private)
		VALUES ($1, $2, '', 'dm', true)
		RETURNING id, workspace_id, name, type, is_private, created_at
	`, channelID, workspaceID).Scan(
		&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Type, &ch.IsPrivate, &ch.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("dm: insert channel: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO dm_channels (channel_id, user1_id, user2_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user1_id, user2_id) DO NOTHING
	`, channelID, lo, hi)
	if err != nil {
		return nil, fmt.Errorf("dm: insert dm_channels: %w", err)
	}

	// Проверяем не выиграл ли другой поток (ON CONFLICT DO NOTHING мог проигнорировать вставку)
	var actualChannelID string
	err = tx.QueryRow(ctx, `
		SELECT channel_id FROM dm_channels WHERE user1_id = $1 AND user2_id = $2
	`, lo, hi).Scan(&actualChannelID)
	if err != nil {
		return nil, fmt.Errorf("dm: recheck channel: %w", err)
	}

	if actualChannelID != channelID {
		// Другой поток успел создать — откатываем и возвращаем существующий
		_ = tx.Rollback(ctx)
		return r.getChannelByID(ctx, actualChannelID)
	}

	// Добавляем обоих участников в channel_members
	for _, uid := range []string{userID1, userID2} {
		_, err = tx.Exec(ctx, `
			INSERT INTO channel_members (user_id, channel_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, uid, channelID)
		if err != nil {
			return nil, fmt.Errorf("dm: insert channel_member %s: %w", uid, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("dm: commit: %w", err)
	}

	return ch, nil
}

// GetByUserID возвращает все DM-диалоги пользователя с данными собеседника.
func (r *DMChannelRepo) GetByUserID(ctx context.Context, workspaceID, userID string) ([]*domain.DMChannel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			dc.channel_id,
			c.workspace_id,
			c.created_at,
			u.id             AS peer_user_id,
			u.display_name   AS peer_display_name,
			u.avatar_url     AS peer_avatar_url,
			COALESCE(u.status, 'offline') AS peer_status,
			COALESCE((
				SELECT COUNT(*)
				FROM messages m
				WHERE m.channel_id = dc.channel_id
				  AND m.parent_id IS NULL
				  AND m.user_id != $1
				  AND (
					cm.last_read_message_id IS NULL
					OR m.created_at > (
						SELECT created_at FROM messages WHERE id = cm.last_read_message_id
					)
				  )
			), 0) AS unread_count,
			(
				SELECT MAX(m2.created_at)
				FROM messages m2
				WHERE m2.channel_id = dc.channel_id AND m2.parent_id IS NULL
			) AS last_message_at
		FROM dm_channels dc
		JOIN channels c ON c.id = dc.channel_id
		JOIN users u ON u.id = CASE
			WHEN dc.user1_id = $1 THEN dc.user2_id
			ELSE dc.user1_id
		END
		LEFT JOIN channel_members cm ON cm.channel_id = dc.channel_id AND cm.user_id = $1
		WHERE c.workspace_id = $2
		  AND (dc.user1_id = $1 OR dc.user2_id = $1)
		ORDER BY last_message_at DESC NULLS LAST, c.created_at DESC
	`, userID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("dm: list by user: %w", err)
	}
	defer rows.Close()

	var result []*domain.DMChannel
	for rows.Next() {
		dm := &domain.DMChannel{}
		if err := rows.Scan(
			&dm.ChannelID,
			&dm.WorkspaceID,
			&dm.CreatedAt,
			&dm.Peer.UserID,
			&dm.Peer.DisplayName,
			&dm.Peer.AvatarURL,
			&dm.Peer.Status,
			&dm.UnreadCount,
			&dm.LastMessageAt,
		); err != nil {
			return nil, fmt.Errorf("dm: scan row: %w", err)
		}
		result = append(result, dm)
	}
	return result, rows.Err()
}

// IsMember проверяет что пользователь является участником DM-канала.
func (r *DMChannelRepo) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM channel_members
			WHERE channel_id = $1 AND user_id = $2
		)
	`, channelID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("dm: is member: %w", err)
	}
	return exists, nil
}

func (r *DMChannelRepo) findExisting(ctx context.Context, workspaceID, lo, hi string) (*domain.Channel, error) {
	var channelID string
	err := r.db.QueryRow(ctx, `
		SELECT dc.channel_id
		FROM dm_channels dc
		JOIN channels c ON c.id = dc.channel_id
		WHERE dc.user1_id = $1 AND dc.user2_id = $2 AND c.workspace_id = $3
	`, lo, hi, workspaceID).Scan(&channelID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.getChannelByID(ctx, channelID)
}

func (r *DMChannelRepo) getChannelByID(ctx context.Context, id string) (*domain.Channel, error) {
	ch := &domain.Channel{}
	err := r.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, type, is_private, created_at
		FROM channels WHERE id = $1
	`, id).Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Type, &ch.IsPrivate, &ch.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("dm: get channel by id: %w", err)
	}
	return ch, nil
}
