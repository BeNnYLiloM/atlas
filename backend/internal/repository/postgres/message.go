package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

// ErrCallStatusTransitionDenied — переход статуса звонка запрещён (неверный текущий статус или нет прав).
var ErrCallStatusTransitionDenied = errors.New("call status transition denied")

type MessageRepo struct {
	db *pgxpool.Pool
}

func NewMessageRepo(db *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (id, channel_id, user_id, content, parent_id, type, call_status, call_duration_sec)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		message.ID,
		message.ChannelID,
		message.UserID,
		message.Content,
		message.ParentID,
		message.Type,
		message.CallStatus,
		message.CallDurationSec,
	).Scan(&message.CreatedAt)
}

// UpdateCallStatus обновляет статус и длительность call-сообщения.
// callerID — user_id создателя звонка: только он может обновить запись.
// allowedFromStatuses — список текущих статусов при которых обновление разрешено (защита от race condition).
func (r *MessageRepo) UpdateCallStatus(ctx context.Context, messageID, callerID, status string, durationSec *int, allowedFromStatuses []string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE messages
		SET call_status = $3, call_duration_sec = $4, updated_at = NOW()
		WHERE id = $1
		  AND user_id = $2
		  AND type = 'call'
		  AND call_status = ANY($5)
	`, messageID, callerID, status, durationSec, allowedFromStatuses)
	if err != nil {
		return err
	}
	// Если 0 строк — либо сообщение не найдено, либо переход статуса недопустим
	if tag.RowsAffected() == 0 {
		return ErrCallStatusTransitionDenied
	}
	return nil
}

func (r *MessageRepo) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.parent_id,
		       m.type, m.call_status, m.call_duration_sec,
		       m.created_at, m.updated_at,
		       u.id, u.display_name, u.avatar_url
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`
	msg := &domain.Message{}
	author := &domain.MessageAuthor{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&msg.ID,
		&msg.ChannelID,
		&msg.UserID,
		&msg.Content,
		&msg.ParentID,
		&msg.Type,
		&msg.CallStatus,
		&msg.CallDurationSec,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&author.ID,
		&author.DisplayName,
		&author.AvatarURL,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	msg.User = author
	return msg, nil
}

func (r *MessageRepo) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.parent_id,
		       m.type, m.call_status, m.call_duration_sec,
		       m.created_at, m.updated_at,
		       u.id, u.display_name, u.avatar_url,
		       COALESCE((SELECT COUNT(*) FROM messages WHERE parent_id = m.id), 0) AS thread_replies_count
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
		author := &domain.MessageAuthor{}
		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.ParentID,
			&msg.Type, &msg.CallStatus, &msg.CallDurationSec,
			&msg.CreatedAt, &msg.UpdatedAt,
			&author.ID, &author.DisplayName, &author.AvatarURL,
			&msg.ThreadRepliesCount,
		); err != nil {
			return nil, err
		}
		msg.User = author
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (r *MessageRepo) GetThreadMessages(ctx context.Context, parentID string) ([]*domain.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.content, m.parent_id,
		       m.type, m.call_status, m.call_duration_sec,
		       m.created_at, m.updated_at,
		       u.id, u.display_name, u.avatar_url
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
		author := &domain.MessageAuthor{}
		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.ParentID,
			&msg.Type, &msg.CallStatus, &msg.CallDurationSec,
			&msg.CreatedAt, &msg.UpdatedAt,
			&author.ID, &author.DisplayName, &author.AvatarURL,
		); err != nil {
			return nil, err
		}
		msg.User = author
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (r *MessageRepo) Update(ctx context.Context, message *domain.Message) error {
	_, err := r.db.Exec(ctx, `
		UPDATE messages
		SET content = $2, updated_at = NOW()
		WHERE id = $1
	`, message.ID, message.Content)
	return err
}

func (r *MessageRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM messages WHERE id = $1", id)
	return err
}
