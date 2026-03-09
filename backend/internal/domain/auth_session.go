package domain

import "time"

type AuthSession struct {
	ID                  string     `json:"id" db:"id"`
	FamilyID            string     `json:"family_id" db:"family_id"`
	UserID              string     `json:"user_id" db:"user_id"`
	RefreshTokenHash    string     `json:"-" db:"refresh_token_hash"`
	UserAgent           string     `json:"user_agent" db:"user_agent"`
	IPAddress           string     `json:"ip_address" db:"ip_address"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt           time.Time  `json:"expires_at" db:"expires_at"`
	LastUsedAt          time.Time  `json:"last_used_at" db:"last_used_at"`
	RevokedAt           *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	ReplacedBySessionID *string    `json:"replaced_by_session_id,omitempty" db:"replaced_by_session_id"`
}
