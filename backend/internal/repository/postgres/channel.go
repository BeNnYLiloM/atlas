package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type ChannelRepo struct {
	db *pgxpool.Pool
}

func NewChannelRepo(db *pgxpool.Pool) *ChannelRepo {
	return &ChannelRepo{db: db}
}

func (r *ChannelRepo) Create(ctx context.Context, channel *domain.Channel) error {
	query := `
		INSERT INTO channels (id, workspace_id, name, type, is_private, category_id, project_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		channel.ID,
		channel.WorkspaceID,
		channel.Name,
		channel.Type,
		channel.IsPrivate,
		channel.CategoryID,
		channel.ProjectID,
	).Scan(&channel.CreatedAt)
}

func (r *ChannelRepo) GetByID(ctx context.Context, id string) (*domain.Channel, error) {
	query := `
		SELECT id, workspace_id, name, type, is_private, topic, slowmode_seconds,
		       position, category_id, project_id, created_at
		FROM channels WHERE id = $1
	`
	ch := &domain.Channel{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Type,
		&ch.IsPrivate, &ch.Topic, &ch.SlowmodeSeconds,
		&ch.Position, &ch.CategoryID, &ch.ProjectID, &ch.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return ch, err
}

func (r *ChannelRepo) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.Channel, error) {
	query := `
		SELECT id, workspace_id, name, type, is_private, topic, slowmode_seconds,
		       position, category_id, project_id, created_at
		FROM channels WHERE workspace_id = $1
		ORDER BY position ASC, created_at ASC
	`
	return r.scanChannels(ctx, query, workspaceID)
}

func (r *ChannelRepo) GetByProjectID(ctx context.Context, projectID string) ([]*domain.Channel, error) {
	query := `
		SELECT id, workspace_id, name, type, is_private, topic, slowmode_seconds,
		       position, category_id, project_id, created_at
		FROM channels WHERE project_id = $1
		ORDER BY position ASC, created_at ASC
	`
	return r.scanChannels(ctx, query, projectID)
}

// GetVisibleByProjectID — каналы проекта видимые пользователю:
// публичные + приватные где у него явный доступ через channel_user_permissions или channel_role_permissions
func (r *ChannelRepo) GetVisibleByProjectID(ctx context.Context, projectID, userID string, roleIDs []string) ([]*domain.Channel, error) {
	query := `
		SELECT DISTINCT c.id, c.workspace_id, c.name, c.type, c.is_private,
		       c.topic, c.slowmode_seconds, c.position, c.category_id, c.project_id, c.created_at
		FROM channels c
		WHERE c.project_id = $1
		  AND (
		    c.is_private = FALSE
		    OR EXISTS (
		      SELECT 1 FROM channel_user_permissions cup
		      WHERE cup.channel_id = c.id AND cup.user_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM channel_role_permissions crp
		      WHERE crp.channel_id = c.id AND crp.role_id = ANY($3)
		    )
		    OR EXISTS (
		      SELECT 1 FROM channel_members cm
		      WHERE cm.channel_id = c.id AND cm.user_id = $2
		    )
		  )
		ORDER BY c.position ASC, c.created_at ASC
	`
	return r.scanChannels(ctx, query, projectID, userID, roleIDs)
}

func (r *ChannelRepo) GetVisibleByWorkspaceID(ctx context.Context, workspaceID, userID string, roleIDs []string) ([]*domain.Channel, error) {
	query := `
		SELECT DISTINCT c.id, c.workspace_id, c.name, c.type, c.is_private,
		       c.topic, c.slowmode_seconds, c.position, c.category_id, c.project_id, c.created_at
		FROM channels c
		WHERE c.workspace_id = $1
		  AND (
		    c.is_private = FALSE
		    OR EXISTS (
		      SELECT 1 FROM channel_user_permissions cup
		      WHERE cup.channel_id = c.id AND cup.user_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM channel_role_permissions crp
		      WHERE crp.channel_id = c.id AND crp.role_id = ANY($3)
		    )
		  )
		ORDER BY c.position ASC, c.created_at ASC
	`
	return r.scanChannels(ctx, query, workspaceID, userID, roleIDs)
}

func (r *ChannelRepo) Update(ctx context.Context, id string, update *domain.ChannelUpdate) (*domain.Channel, error) {
	query := `
		UPDATE channels
		SET
			name             = COALESCE($2, name),
			topic            = COALESCE($3, topic),
			is_private       = COALESCE($4, is_private),
			slowmode_seconds = COALESCE($5, slowmode_seconds),
			category_id      = CASE WHEN $6::boolean THEN $7::uuid ELSE category_id END
		WHERE id = $1
		RETURNING id, workspace_id, name, type, is_private, topic, slowmode_seconds,
		          position, category_id, project_id, created_at
	`
	setCategoryID := update.CategoryID != nil
	var categoryIDVal *string
	if setCategoryID {
		categoryIDVal = update.CategoryID
	}
	ch := &domain.Channel{}
	err := r.db.QueryRow(ctx, query,
		id, update.Name, update.Topic, update.IsPrivate, update.SlowmodeSeconds,
		setCategoryID, categoryIDVal,
	).Scan(
		&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Type,
		&ch.IsPrivate, &ch.Topic, &ch.SlowmodeSeconds,
		&ch.Position, &ch.CategoryID, &ch.ProjectID, &ch.CreatedAt,
	)
	return ch, err
}

func (r *ChannelRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM channels WHERE id = $1", id)
	return err
}

func (r *ChannelRepo) scanChannels(ctx context.Context, query string, args ...interface{}) ([]*domain.Channel, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*domain.Channel
	for rows.Next() {
		ch := &domain.Channel{}
		if err := rows.Scan(
			&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Type,
			&ch.IsPrivate, &ch.Topic, &ch.SlowmodeSeconds,
			&ch.Position, &ch.CategoryID, &ch.ProjectID, &ch.CreatedAt,
		); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
