package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type MessageRepo struct {
	db *pgxpool.Pool
}

func NewMessageRepo(db *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (id, channel_id, user_id, content, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		message.ID,
		message.ChannelID,
		message.UserID,
		message.Content,
		message.ParentID,
	).Scan(&message.CreatedAt)
}

func (r *MessageRepo) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.parent_id, m.created_at, m.updated_at,
			   u.id, u.email, u.display_name, u.avatar_url, u.created_at
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`
	msg := &domain.Message{}
	user := &domain.User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&msg.ID,
		&msg.ChannelID,
		&msg.UserID,
		&msg.Content,
		&msg.ParentID,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	msg.User = user
	return msg, nil
}

func (r *MessageRepo) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.parent_id, m.created_at, m.updated_at,
			   u.id, u.email, u.display_name, u.avatar_url, u.created_at,
			   COALESCE((SELECT COUNT(*) FROM messages WHERE parent_id = m.id), 0) as thread_replies_count
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.channel_id = $1 AND m.parent_id IS NULL
		ORDER BY m.created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		user := &domain.User{}
		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.ParentID, &msg.CreatedAt, &msg.UpdatedAt,
			&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.CreatedAt,
			&msg.ThreadRepliesCount,
		); err != nil {
			return nil, err
		}
		msg.User = user
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (r *MessageRepo) GetThreadMessages(ctx context.Context, parentID string) ([]*domain.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.parent_id, m.created_at, m.updated_at,
			   u.id, u.email, u.display_name, u.avatar_url, u.created_at
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.parent_id = $1
		ORDER BY m.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		user := &domain.User{}
		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.ParentID, &msg.CreatedAt, &msg.UpdatedAt,
			&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.CreatedAt,
		); err != nil {
			return nil, err
		}
		msg.User = user
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (r *MessageRepo) Update(ctx context.Context, message *domain.Message) error {
	query := `
		UPDATE messages 
		SET content = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, message.ID, message.Content)
	return err
}

func (r *MessageRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM messages WHERE id = $1", id)
	return err
}

