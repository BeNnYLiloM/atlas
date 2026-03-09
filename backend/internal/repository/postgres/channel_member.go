package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type ChannelMemberRepo struct {
	db *pgxpool.Pool
}

func NewChannelMemberRepo(db *pgxpool.Pool) *ChannelMemberRepo {
	return &ChannelMemberRepo{db: db}
}

// UpsertMember добавляет или обновляет участника канала
func (r *ChannelMemberRepo) UpsertMember(ctx context.Context, userID, channelID string) error {
	query := `
		INSERT INTO channel_members (user_id, channel_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, channel_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userID, channelID)
	return err
}

// MarkAsRead отмечает канал прочитанным до определенного сообщения
func (r *ChannelMemberRepo) MarkAsRead(ctx context.Context, userID, channelID string, messageID *string) error {
	query := `
		INSERT INTO channel_members (user_id, channel_id, last_read_message_id, last_read_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, channel_id) 
		DO UPDATE SET 
			last_read_message_id = EXCLUDED.last_read_message_id,
			last_read_at = EXCLUDED.last_read_at
	`
	_, err := r.db.Exec(ctx, query, userID, channelID, messageID)
	return err
}

// GetUnreadCount возвращает количество непрочитанных сообщений в канале для пользователя
func (r *ChannelMemberRepo) GetUnreadCount(ctx context.Context, userID, channelID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		LEFT JOIN channel_members cm ON cm.channel_id = m.channel_id AND cm.user_id = $1
		WHERE m.channel_id = $2
		  AND m.parent_id IS NULL
		  AND (
		    cm.last_read_message_id IS NULL 
		    OR m.created_at > (
		      SELECT created_at FROM messages WHERE id = cm.last_read_message_id
		    )
		  )
		  AND m.user_id != $1
	`
	var count int
	err := r.db.QueryRow(ctx, query, userID, channelID).Scan(&count)
	return count, err
}

// GetUnreadCountsForWorkspace возвращает количество непрочитанных, упоминаний и уровень уведомлений для всех каналов workspace
func (r *ChannelMemberRepo) GetUnreadCountsForWorkspace(ctx context.Context, userID, workspaceID string) (map[string]domain.ChannelStats, error) {
	query := `
		WITH user_name AS (
			SELECT display_name FROM users WHERE id = $1
		),
		last_read AS (
			SELECT channel_id, last_read_message_id
			FROM channel_members
			WHERE user_id = $1
		),
		unread_msgs AS (
			SELECT
				m.channel_id,
				m.id,
				m.content
			FROM messages m
			LEFT JOIN last_read lr ON lr.channel_id = m.channel_id
			WHERE m.parent_id IS NULL
			  AND m.user_id != $1
			  AND (
				lr.last_read_message_id IS NULL
				OR m.created_at > (
					SELECT created_at FROM messages WHERE id = lr.last_read_message_id
				)
			  )
		)
		SELECT
			c.id AS channel_id,
			COALESCE(COUNT(um.id), 0) AS unread_count,
			COALESCE(COUNT(um.id) FILTER (
				WHERE um.content LIKE '%@' || (SELECT display_name FROM user_name) || '%'
				   OR um.content LIKE '%@everyone%'
			), 0) AS mention_count,
			COALESCE(cm.notification_level, 'all') AS notification_level
		FROM channels c
		LEFT JOIN channel_members cm ON cm.channel_id = c.id AND cm.user_id = $1
		LEFT JOIN unread_msgs um ON um.channel_id = c.id
		WHERE c.workspace_id = $2
		GROUP BY c.id, cm.notification_level
	`

	rows, err := r.db.Query(ctx, query, userID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]domain.ChannelStats)
	for rows.Next() {
		var channelID string
		var s domain.ChannelStats
		if err := rows.Scan(&channelID, &s.UnreadCount, &s.MentionCount, &s.NotificationLevel); err != nil {
			return nil, err
		}
		stats[channelID] = s
	}

	return stats, rows.Err()
}

// GetLastReadMessageID возвращает ID последнего прочитанного сообщения
func (r *ChannelMemberRepo) GetLastReadMessageID(ctx context.Context, userID, channelID string) (*string, error) {
	query := `
		SELECT last_read_message_id 
		FROM channel_members 
		WHERE user_id = $1 AND channel_id = $2
	`
	var messageID *string
	err := r.db.QueryRow(ctx, query, userID, channelID).Scan(&messageID)
	if err != nil {
		return nil, nil
	}
	return messageID, err
}

// GetMembers возвращает список участников канала
func (r *ChannelMemberRepo) GetMembers(ctx context.Context, channelID string) ([]*domain.ChannelMemberInfo, error) {
	query := `
		SELECT cm.user_id, cm.channel_id, u.display_name, u.avatar_url
		FROM channel_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.channel_id = $1
		ORDER BY u.display_name
	`
	rows, err := r.db.Query(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.ChannelMemberInfo
	for rows.Next() {
		m := &domain.ChannelMemberInfo{}
		if err := rows.Scan(&m.UserID, &m.ChannelID, &m.DisplayName, &m.AvatarURL); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// RemoveMember удаляет участника из канала
func (r *ChannelMemberRepo) RemoveMember(ctx context.Context, userID, channelID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM channel_members WHERE user_id = $1 AND channel_id = $2`,
		userID, channelID,
	)
	return err
}

// UpdateNotificationLevel обновляет уровень уведомлений для пользователя в канале
func (r *ChannelMemberRepo) UpdateNotificationLevel(ctx context.Context, userID, channelID, level string) error {
	query := `
		INSERT INTO channel_members (user_id, channel_id, notification_level)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, channel_id)
		DO UPDATE SET notification_level = $3
	`
	_, err := r.db.Exec(ctx, query, userID, channelID, level)
	return err
}

// GetNotificationLevel возвращает уровень уведомлений пользователя в канале
func (r *ChannelMemberRepo) GetNotificationLevel(ctx context.Context, userID, channelID string) (string, error) {
	var level string
	err := r.db.QueryRow(ctx,
		`SELECT notification_level FROM channel_members WHERE user_id = $1 AND channel_id = $2`,
		userID, channelID,
	).Scan(&level)
	if err != nil {
		return domain.NotificationAll, nil
	}
	return level, nil
}

// GetLastMessageAt возвращает время последнего сообщения пользователя в канале
func (r *ChannelMemberRepo) GetLastMessageAt(ctx context.Context, userID, channelID string) (*time.Time, error) {
	var t *time.Time
	err := r.db.QueryRow(ctx,
		`SELECT last_message_at FROM channel_members WHERE user_id = $1 AND channel_id = $2`,
		userID, channelID,
	).Scan(&t)
	if err != nil {
		return nil, nil
	}
	return t, nil
}

// SetLastMessageAt обновляет время последнего сообщения пользователя в канале
func (r *ChannelMemberRepo) SetLastMessageAt(ctx context.Context, userID, channelID string) error {
	query := `
		INSERT INTO channel_members (user_id, channel_id, last_message_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, channel_id)
		DO UPDATE SET last_message_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, userID, channelID)
	return err
}

// MarkThreadAsRead отмечает тред прочитанным до определенного сообщения
func (r *ChannelMemberRepo) MarkThreadAsRead(ctx context.Context, userID, parentMessageID string, lastMessageID *string) error {
	// Получаем channel_id родительского сообщения
	var channelID string
	err := r.db.QueryRow(ctx, `SELECT channel_id FROM messages WHERE id = $1`, parentMessageID).Scan(&channelID)
	if err != nil {
		return err
	}

	// Если lastMessageID не указан, удаляем запись о прочтении треда
	if lastMessageID == nil || *lastMessageID == "" {
		query := `
			UPDATE channel_members
			SET thread_last_reads = COALESCE(thread_last_reads, '{}'::jsonb) - $3::text
			WHERE user_id = $1 AND channel_id = $2
		`
		_, err = r.db.Exec(ctx, query, userID, channelID, parentMessageID)
		return err
	}

	// Обновляем thread_last_reads JSONB поле
	query := `
		INSERT INTO channel_members (user_id, channel_id, thread_last_reads)
		VALUES ($1, $2, jsonb_build_object($3::text, $4::text))
		ON CONFLICT (user_id, channel_id) 
		DO UPDATE SET 
			thread_last_reads = COALESCE(channel_members.thread_last_reads, '{}'::jsonb) || jsonb_build_object($3::text, $4::text)
	`
	_, err = r.db.Exec(ctx, query, userID, channelID, parentMessageID, *lastMessageID)
	return err
}

// GetThreadUnreadCount возвращает количество непрочитанных сообщений в треде
func (r *ChannelMemberRepo) GetThreadUnreadCount(ctx context.Context, userID, parentMessageID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		LEFT JOIN (
			SELECT 
				cm.user_id,
				(cm.thread_last_reads->>$2::text)::uuid as last_read_msg_id
			FROM channel_members cm
			JOIN messages pm ON pm.channel_id = cm.channel_id
			WHERE cm.user_id = $1 AND pm.id = $2::uuid
		) cm ON true
		WHERE m.parent_id = $2::uuid
		  AND m.user_id != $1
		  AND (
		    cm.last_read_msg_id IS NULL 
		    OR m.created_at > (
		      SELECT created_at FROM messages WHERE id = cm.last_read_msg_id
		    )
		  )
	`
	var count int
	err := r.db.QueryRow(ctx, query, userID, parentMessageID).Scan(&count)
	return count, err
}

