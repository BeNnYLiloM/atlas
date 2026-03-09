package domain

import "time"

// ChannelMember представляет участника канала с информацией о прочитанных сообщениях
type ChannelMember struct {
	UserID            string     `json:"user_id" db:"user_id"`
	ChannelID         string     `json:"channel_id" db:"channel_id"`
	LastReadMessageID *string    `json:"last_read_message_id,omitempty" db:"last_read_message_id"`
	LastReadAt        *time.Time `json:"last_read_at,omitempty" db:"last_read_at"`
	JoinedAt          time.Time  `json:"joined_at" db:"joined_at"`
	NotificationLevel string     `json:"notification_level" db:"notification_level"` // all | mentions | nothing
}

const (
	NotificationAll      = "all"
	NotificationMentions = "mentions"
	NotificationNothing  = "nothing"
)

type UpdateNotificationsInput struct {
	Level string `json:"level" binding:"required,oneof=all mentions nothing"`
}

// ChannelMemberInfo — участник канала с публичной информацией пользователя
type ChannelMemberInfo struct {
	UserID      string  `json:"user_id" db:"user_id"`
	ChannelID   string  `json:"channel_id" db:"channel_id"`
	DisplayName string  `json:"display_name" db:"display_name"`
	AvatarURL   *string `json:"avatar_url" db:"avatar_url"`
}

// ChannelWithUnread - канал с количеством непрочитанных сообщений
type ChannelWithUnread struct {
	*Channel
	UnreadCount       int    `json:"unread_count"`
	MentionCount      int    `json:"mention_count"`
	NotificationLevel string `json:"notification_level"`
}

// ChannelStats - агрегированная статистика канала для пользователя
type ChannelStats struct {
	UnreadCount       int
	MentionCount      int
	NotificationLevel string
}

// MarkAsReadInput - данные для отметки канала прочитанным
type MarkAsReadInput struct {
	MessageID *string `json:"message_id,omitempty"` // ID последнего прочитанного сообщения (опционально)
}
