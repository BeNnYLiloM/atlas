package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type AuthSessionRepo struct {
	db *pgxpool.Pool
}

func NewAuthSessionRepo(db *pgxpool.Pool) *AuthSessionRepo {
	return &AuthSessionRepo{db: db}
}

func (r *AuthSessionRepo) Create(ctx context.Context, session *domain.AuthSession) error {
	query := `
		INSERT INTO auth_sessions (
			id, family_id, user_id, refresh_token_hash, user_agent, ip_address,
			created_at, expires_at, last_used_at, revoked_at, replaced_by_session_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.FamilyID,
		session.UserID,
		session.RefreshTokenHash,
		session.UserAgent,
		session.IPAddress,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastUsedAt,
		session.RevokedAt,
		session.ReplacedBySessionID,
	)
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.Create: %w", err)
	}

	return nil
}

func (r *AuthSessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.AuthSession, error) {
	query := `
		SELECT id, family_id, user_id, refresh_token_hash, user_agent, ip_address,
		       created_at, expires_at, last_used_at, revoked_at, replaced_by_session_id
		FROM auth_sessions
		WHERE refresh_token_hash = $1
	`

	session := &domain.AuthSession{}
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&session.ID,
		&session.FamilyID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastUsedAt,
		&session.RevokedAt,
		&session.ReplacedBySessionID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("AuthSessionRepo.GetByRefreshTokenHash: %w", err)
	}

	return session, nil
}

func (r *AuthSessionRepo) Rotate(ctx context.Context, currentSessionID string, nextSession *domain.AuthSession) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.Rotate begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Сначала создаем новую сессию, чтобы foreign key на replaced_by_session_id был валиден.
	_, err = tx.Exec(ctx, `
		INSERT INTO auth_sessions (
			id, family_id, user_id, refresh_token_hash, user_agent, ip_address,
			created_at, expires_at, last_used_at, revoked_at, replaced_by_session_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		nextSession.ID,
		nextSession.FamilyID,
		nextSession.UserID,
		nextSession.RefreshTokenHash,
		nextSession.UserAgent,
		nextSession.IPAddress,
		nextSession.CreatedAt,
		nextSession.ExpiresAt,
		nextSession.LastUsedAt,
		nextSession.RevokedAt,
		nextSession.ReplacedBySessionID,
	)
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.Rotate insert next: %w", err)
	}

	cmd, err := tx.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = $2, replaced_by_session_id = $3
		WHERE id = $1 AND revoked_at IS NULL
	`, currentSessionID, nextSession.CreatedAt, nextSession.ID)
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.Rotate revoke current: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("AuthSessionRepo.Rotate commit: %w", err)
	}

	return nil
}

func (r *AuthSessionRepo) RevokeByID(ctx context.Context, sessionID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE id = $1
	`, sessionID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.RevokeByID: %w", err)
	}

	return nil
}

func (r *AuthSessionRepo) RevokeByIDForUser(ctx context.Context, sessionID, userID string) (bool, error) {
	cmd, err := r.db.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $3)
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, sessionID, userID, time.Now().UTC())
	if err != nil {
		return false, fmt.Errorf("AuthSessionRepo.RevokeByIDForUser: %w", err)
	}
	return cmd.RowsAffected() == 1, nil
}

func (r *AuthSessionRepo) RevokeFamily(ctx context.Context, familyID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE family_id = $1
	`, familyID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.RevokeFamily: %w", err)
	}

	return nil
}

func (r *AuthSessionRepo) ListActiveByUserID(ctx context.Context, userID string) ([]*domain.AuthSession, error) {
	query := `
		SELECT id, family_id, user_id, refresh_token_hash, user_agent, ip_address,
		       created_at, expires_at, last_used_at, revoked_at, replaced_by_session_id
		FROM auth_sessions
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		ORDER BY last_used_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("AuthSessionRepo.ListActiveByUserID: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.AuthSession
	for rows.Next() {
		s := &domain.AuthSession{}
		if err := rows.Scan(
			&s.ID, &s.FamilyID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress,
			&s.CreatedAt, &s.ExpiresAt, &s.LastUsedAt, &s.RevokedAt, &s.ReplacedBySessionID,
		); err != nil {
			return nil, fmt.Errorf("AuthSessionRepo.ListActiveByUserID scan: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *AuthSessionRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("AuthSessionRepo.RevokeAllByUserID: %w", err)
	}

	return nil
}
