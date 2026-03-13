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

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
	).Scan(&user.CreatedAt)
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, avatar_url, status, custom_status, last_seen, created_at
		FROM users WHERE id = $1
	`
	return r.scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *UserRepo) GetStatusByID(ctx context.Context, userID string) (string, error) {
	var status string
	err := r.db.QueryRow(ctx, `SELECT status FROM users WHERE id = $1`, userID).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("UserRepo.GetStatusByID: %w", err)
	}
	return status, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, avatar_url, status, custom_status, last_seen, created_at
		FROM users WHERE email = $1
	`
	return r.scanUser(r.db.QueryRow(ctx, query, email))
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET display_name = $2, avatar_url = $3, status = $4, custom_status = $5
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.DisplayName,
		user.AvatarURL,
		user.Status,
		user.CustomStatus,
	)
	if err != nil {
		return fmt.Errorf("UserRepo.Update: %w", err)
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, newHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET password_hash = $2 WHERE id = $1
	`, userID, newHash)
	if err != nil {
		return fmt.Errorf("UserRepo.UpdatePassword: %w", err)
	}
	return nil
}

func (r *UserRepo) UpdateLastSeen(ctx context.Context, userID string, t time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET last_seen = $2 WHERE id = $1
	`, userID, t)
	if err != nil {
		return fmt.Errorf("UserRepo.UpdateLastSeen: %w", err)
	}
	return nil
}

func (r *UserRepo) DeleteByID(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("UserRepo.DeleteByID: %w", err)
	}
	return nil
}

func (r *UserRepo) scanUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Status,
		&user.CustomStatus,
		&user.LastSeen,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("UserRepo.scanUser: %w", err)
	}
	return user, nil
}
