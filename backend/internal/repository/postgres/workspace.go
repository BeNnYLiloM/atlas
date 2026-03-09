package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type WorkspaceRepo struct {
	db *pgxpool.Pool
}

func NewWorkspaceRepo(db *pgxpool.Pool) *WorkspaceRepo {
	return &WorkspaceRepo{db: db}
}

func (r *WorkspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, owner_id)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.OwnerID,
	).Scan(&workspace.CreatedAt)
}

func (r *WorkspaceRepo) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	query := `
		SELECT id, name, owner_id, description, icon_url, created_at
		FROM workspaces WHERE id = $1
	`
	ws := &domain.Workspace{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ws.ID,
		&ws.Name,
		&ws.OwnerID,
		&ws.Description,
		&ws.IconURL,
		&ws.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (r *WorkspaceRepo) Update(ctx context.Context, id string, update *domain.WorkspaceUpdate) (*domain.Workspace, error) {
	query := `
		UPDATE workspaces
		SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			icon_url    = COALESCE($4, icon_url)
		WHERE id = $1
		RETURNING id, name, owner_id, description, icon_url, created_at
	`
	ws := &domain.Workspace{}
	err := r.db.QueryRow(ctx, query, id, update.Name, update.Description, update.IconURL).Scan(
		&ws.ID,
		&ws.Name,
		&ws.OwnerID,
		&ws.Description,
		&ws.IconURL,
		&ws.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (r *WorkspaceRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	query := `
		SELECT w.id, w.name, w.owner_id, w.description, w.icon_url, w.created_at
		FROM workspaces w
		JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE wm.user_id = $1
		ORDER BY w.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		ws := &domain.Workspace{}
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.Description, &ws.IconURL, &ws.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, rows.Err()
}

func (r *WorkspaceRepo) AddMember(ctx context.Context, member *domain.WorkspaceMember) error {
	query := `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = $3
	`
	_, err := r.db.Exec(ctx, query, member.WorkspaceID, member.UserID, member.Role)
	return err
}

func (r *WorkspaceRepo) GetMembers(ctx context.Context, workspaceID string) ([]*domain.WorkspaceMember, error) {
	query := `
		SELECT
			wm.workspace_id,
			wm.user_id,
			wm.role,
			u.display_name,
			u.avatar_url,
			wm.nickname,
			COALESCE(
				json_agg(
					json_build_object(
						'id', wr.id,
						'workspace_id', wr.workspace_id,
						'name', wr.name,
						'color', wr.color,
						'position', wr.position,
						'is_system', wr.is_system,
						'permissions', wr.permissions,
						'created_at', wr.created_at
					) ORDER BY wr.position DESC
				) FILTER (WHERE wr.id IS NOT NULL),
				'[]'
			) AS custom_roles
		FROM workspace_members wm
		JOIN users u ON u.id = wm.user_id
		LEFT JOIN workspace_member_roles wmr ON wmr.workspace_id = wm.workspace_id AND wmr.user_id = wm.user_id
		LEFT JOIN workspace_roles wr ON wr.id = wmr.role_id AND wr.is_system = FALSE
		WHERE wm.workspace_id = $1
		GROUP BY wm.workspace_id, wm.user_id, wm.role, u.display_name, u.avatar_url, wm.nickname
		ORDER BY
			CASE wm.role
				WHEN 'owner' THEN 1
				WHEN 'admin' THEN 2
				WHEN 'member' THEN 3
				ELSE 4
			END,
			u.display_name
	`
	rows, err := r.db.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.WorkspaceMember
	for rows.Next() {
		m := &domain.WorkspaceMember{}
		var rolesJSON []byte
		if err := rows.Scan(&m.WorkspaceID, &m.UserID, &m.Role, &m.DisplayName, &m.AvatarURL, &m.Nickname, &rolesJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rolesJSON, &m.CustomRoles); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *WorkspaceRepo) GetMemberUserIDs(ctx context.Context, workspaceID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM workspace_members WHERE workspace_id = $1`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *WorkspaceRepo) UpdateMember(ctx context.Context, workspaceID, userID string, update *domain.WorkspaceMemberUpdate) error {
	query := `
		UPDATE workspace_members
		SET
			role     = COALESCE($3, role),
			nickname = COALESCE($4, nickname)
		WHERE workspace_id = $1 AND user_id = $2
	`
	_, err := r.db.Exec(ctx, query, workspaceID, userID, update.Role, update.Nickname)
	return err
}

func (r *WorkspaceRepo) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID)
	return err
}

func (r *WorkspaceRepo) GetMember(ctx context.Context, workspaceID, userID string) (*domain.WorkspaceMember, error) {
	query := `
		SELECT workspace_id, user_id, role
		FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`
	m := &domain.WorkspaceMember{}
	err := r.db.QueryRow(ctx, query, workspaceID, userID).Scan(
		&m.WorkspaceID,
		&m.UserID,
		&m.Role,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *WorkspaceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM workspaces WHERE id = $1", id)
	return err
}

