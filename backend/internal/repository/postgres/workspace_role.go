package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type WorkspaceRoleRepo struct {
	db *pgxpool.Pool
}

func NewWorkspaceRoleRepo(db *pgxpool.Pool) *WorkspaceRoleRepo {
	return &WorkspaceRoleRepo{db: db}
}

func scanRole(row pgx.Row) (*domain.WorkspaceRole, error) {
	r := &domain.WorkspaceRole{}
	var permJSON []byte
	err := row.Scan(&r.ID, &r.WorkspaceID, &r.Name, &r.Color, &r.Position, &r.IsSystem, &permJSON, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(permJSON, &r.Permissions); err != nil {
		return nil, fmt.Errorf("unmarshal permissions: %w", err)
	}
	return r, nil
}

func (r *WorkspaceRoleRepo) Create(ctx context.Context, role *domain.WorkspaceRole) error {
	permJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, `
		INSERT INTO workspace_roles (id, workspace_id, name, color, position, is_system, permissions)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, name, color, position, is_system, permissions, created_at
	`, role.ID, role.WorkspaceID, role.Name, role.Color, role.Position, role.IsSystem, permJSON).
		Scan(&role.ID, &role.WorkspaceID, &role.Name, &role.Color, &role.Position, &role.IsSystem, &permJSON, &role.CreatedAt)
}

func (r *WorkspaceRoleRepo) GetByID(ctx context.Context, id string) (*domain.WorkspaceRole, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, color, position, is_system, permissions, created_at
		FROM workspace_roles WHERE id = $1
	`, id)
	role, err := scanRole(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return role, err
}

func (r *WorkspaceRoleRepo) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.WorkspaceRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT wr.id, wr.workspace_id, wr.name, wr.color, wr.position, wr.is_system, wr.permissions, wr.created_at
		FROM workspace_roles wr
		WHERE wr.workspace_id = $1
		ORDER BY wr.position DESC, wr.created_at ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.WorkspaceRole
	for rows.Next() {
		role := &domain.WorkspaceRole{}
		var permJSON []byte
		if err := rows.Scan(&role.ID, &role.WorkspaceID, &role.Name, &role.Color, &role.Position, &role.IsSystem, &permJSON, &role.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permJSON, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *WorkspaceRoleRepo) Update(ctx context.Context, id string, update *domain.WorkspaceRoleUpdate) (*domain.WorkspaceRole, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil || existing == nil {
		return nil, err
	}

	if update.Name != nil {
		existing.Name = *update.Name
	}
	if update.Color != nil {
		existing.Color = *update.Color
	}
	if update.Permissions != nil {
		existing.Permissions = *update.Permissions
	}

	permJSON, err := json.Marshal(existing.Permissions)
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, `
		UPDATE workspace_roles
		SET name = $1, color = $2, permissions = $3
		WHERE id = $4 AND is_system = FALSE
		RETURNING id, workspace_id, name, color, position, is_system, permissions, created_at
	`, existing.Name, existing.Color, permJSON, id)
	return scanRole(row)
}

func (r *WorkspaceRoleRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM workspace_roles WHERE id = $1 AND is_system = FALSE
	`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found or is system role")
	}
	return nil
}

func (r *WorkspaceRoleRepo) AssignRole(ctx context.Context, workspaceID, userID, roleID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO workspace_member_roles (workspace_id, user_id, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, workspaceID, userID, roleID)
	return err
}

func (r *WorkspaceRoleRepo) RevokeRole(ctx context.Context, workspaceID, userID, roleID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM workspace_member_roles
		WHERE workspace_id = $1 AND user_id = $2 AND role_id = $3
	`, workspaceID, userID, roleID)
	return err
}

func (r *WorkspaceRoleRepo) GetMemberRoles(ctx context.Context, workspaceID, userID string) ([]*domain.WorkspaceRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT wr.id, wr.workspace_id, wr.name, wr.color, wr.position, wr.is_system, wr.permissions, wr.created_at
		FROM workspace_roles wr
		JOIN workspace_member_roles wmr ON wmr.role_id = wr.id
		WHERE wmr.workspace_id = $1 AND wmr.user_id = $2
		ORDER BY wr.position DESC
	`, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.WorkspaceRole
	for rows.Next() {
		role := &domain.WorkspaceRole{}
		var permJSON []byte
		if err := rows.Scan(&role.ID, &role.WorkspaceID, &role.Name, &role.Color, &role.Position, &role.IsSystem, &permJSON, &role.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permJSON, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *WorkspaceRoleRepo) GetUserIDsByRole(ctx context.Context, roleID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM workspace_member_roles WHERE role_id = $1`, roleID)
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

// GetEffectivePermissions — берёт @everyone + все назначенные роли и объединяет (OR по каждому флагу)
func (r *WorkspaceRoleRepo) GetEffectivePermissions(ctx context.Context, workspaceID, userID string) (*domain.RolePermissions, error) {
	rows, err := r.db.Query(ctx, `
		SELECT wr.permissions
		FROM workspace_roles wr
		WHERE wr.workspace_id = $1
		  AND (
		    wr.name = '@everyone'
		    OR wr.id IN (
		      SELECT role_id FROM workspace_member_roles
		      WHERE workspace_id = $1 AND user_id = $2
		    )
		  )
	`, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &domain.RolePermissions{}
	for rows.Next() {
		var permJSON []byte
		if err := rows.Scan(&permJSON); err != nil {
			return nil, err
		}
		var p domain.RolePermissions
		if err := json.Unmarshal(permJSON, &p); err != nil {
			return nil, err
		}
		// Объединяем: если хоть одна роль даёт право — оно есть
		result.ManageWorkspace = result.ManageWorkspace || p.ManageWorkspace
		result.ManageRoles = result.ManageRoles || p.ManageRoles
		result.ManageChannels = result.ManageChannels || p.ManageChannels
		result.ManageMembers = result.ManageMembers || p.ManageMembers
		result.ViewAuditLog = result.ViewAuditLog || p.ViewAuditLog
		result.SendMessages = result.SendMessages || p.SendMessages
		result.AttachFiles = result.AttachFiles || p.AttachFiles
		result.MentionEveryone = result.MentionEveryone || p.MentionEveryone
		result.ManageMessages = result.ManageMessages || p.ManageMessages
		result.ViewChannels = result.ViewChannels || p.ViewChannels
	}
	return result, rows.Err()
}
