package domain

import (
	"time"
)

const (
	MessageTypeText = "text"
	MessageTypeCall = "call"

	CallStatusRinging   = "ringing"
	CallStatusCancelled = "cancelled"
	CallStatusMissed    = "missed"
	CallStatusOngoing   = "ongoing"
	CallStatusEnded     = "ended"
)

type Message struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Content   string     `json:"content" db:"content"`
	ParentID  *string    `json:"parent_id" db:"parent_id"` // Для тредов
	Type      string     `json:"type" db:"type"`
	// Поля заполнены только для type=call
	CallStatus      *string `json:"call_status,omitempty"       db:"call_status"`
	CallDurationSec *int    `json:"call_duration_sec,omitempty" db:"call_duration_sec"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`

	// Связанные данные (не хранятся в БД)
	User               *MessageAuthor `json:"user,omitempty" db:"-"`
	ThreadRepliesCount int            `json:"thread_replies_count" db:"-"`
	ThreadUnreadCount  int            `json:"thread_unread_count" db:"-"`
}

// MessageAuthor — публичные данные автора сообщения (без email).
type MessageAuthor struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type MessageCreate struct {
	ChannelID string  `json:"channel_id" validate:"required,uuid"`
	Content   string  `json:"content" validate:"required,min=1,max=4000"`
	ParentID  *string `json:"parent_id" validate:"omitempty,uuid"`
}

type MessageUpdate struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}

