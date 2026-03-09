package domain

import (
	"time"
)

type Message struct {
	ID        string     `json:"id" db:"id"`
	ChannelID string     `json:"channel_id" db:"channel_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Content   string     `json:"content" db:"content"`
	ParentID  *string    `json:"parent_id" db:"parent_id"` // Для тредов
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`

	// Связанные данные (не хранятся в БД)
	User               *User  `json:"user,omitempty" db:"-"`
	ThreadRepliesCount int    `json:"thread_replies_count" db:"-"`
	ThreadUnreadCount  int    `json:"thread_unread_count" db:"-"`
}

type MessageCreate struct {
	ChannelID string  `json:"channel_id" validate:"required,uuid"`
	Content   string  `json:"content" validate:"required,min=1,max=4000"`
	ParentID  *string `json:"parent_id" validate:"omitempty,uuid"`
}

type MessageUpdate struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}

