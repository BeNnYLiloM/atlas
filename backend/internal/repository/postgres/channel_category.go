package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type ChannelCategoryRepo struct {
	db *pgxpool.Pool
}

func NewChannelCategoryRepo(db *pgxpool.Pool) *ChannelCategoryRepo {
	return &ChannelCategoryRepo{db: db}
}

func (r *ChannelCategoryRepo) Create(ctx context.Context, cat *domain.ChannelCategory) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO channel_categories (id, workspace_id, name, position, is_private, project_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`, cat.ID, cat.WorkspaceID, cat.Name, cat.Position, cat.IsPrivate, cat.ProjectID).Scan(&cat.CreatedAt)
}

func (r *ChannelCategoryRepo) GetByID(ctx context.Context, id string) (*domain.ChannelCategory, error) {
	cat := &domain.ChannelCategory{}
	err := r.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, position, is_private, project_id, created_at
		FROM channel_categories WHERE id = $1
	`, id).Scan(&cat.ID, &cat.WorkspaceID, &cat.Name, &cat.Position, &cat.IsPrivate, &cat.ProjectID, &cat.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return cat, err
}

func (r *ChannelCategoryRepo) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.ChannelCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, workspace_id, name, position, is_private, project_id, created_at
		FROM channel_categories WHERE workspace_id = $1
		ORDER BY position ASC, created_at ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*domain.ChannelCategory
	for rows.Next() {
		cat := &domain.ChannelCategory{}
		if err := rows.Scan(
			&cat.ID, &cat.WorkspaceID, &cat.Name, &cat.Position,
			&cat.IsPrivate, &cat.ProjectID, &cat.CreatedAt,
		); err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}
	return cats, rows.Err()
}

func (r *ChannelCategoryRepo) Update(ctx context.Context, id string, update *domain.ChannelCategoryUpdate) (*domain.ChannelCategory, error) {
	cat := &domain.ChannelCategory{}
	err := r.db.QueryRow(ctx, `
		UPDATE channel_categories
		SET
			name       = COALESCE($2, name),
			position   = COALESCE($3, position),
			is_private = COALESCE($4, is_private)
		WHERE id = $1
		RETURNING id, workspace_id, name, position, is_private, project_id, created_at
	`, id, update.Name, update.Position, update.IsPrivate).Scan(
		&cat.ID, &cat.WorkspaceID, &cat.Name, &cat.Position,
		&cat.IsPrivate, &cat.ProjectID, &cat.CreatedAt,
	)
	return cat, err
}

func (r *ChannelCategoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM channel_categories WHERE id = $1", id)
	return err
}
