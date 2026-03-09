package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type CategoryPermissionRepo struct {
	db *pgxpool.Pool
}

func NewCategoryPermissionRepo(db *pgxpool.Pool) *CategoryPermissionRepo {
	return &CategoryPermissionRepo{db: db}
}

func (r *CategoryPermissionRepo) GetPermissions(ctx context.Context, categoryID string) (*domain.CategoryPermissions, error) {
	perms := &domain.CategoryPermissions{
		Roles: []domain.ChannelAllowedRole{},
		Users: []domain.ChannelAllowedUser{},
	}

	roleRows, err := r.db.Query(ctx, `
		SELECT crp.category_id, crp.role_id, wr.name, wr.color
		FROM category_role_permissions crp
		JOIN workspace_roles wr ON wr.id = crp.role_id
		WHERE crp.category_id = $1
		ORDER BY wr.position DESC
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		r2 := domain.ChannelAllowedRole{}
		if err := roleRows.Scan(&r2.ChannelID, &r2.RoleID, &r2.RoleName, &r2.RoleColor); err != nil {
			return nil, err
		}
		r2.ChannelID = categoryID
		perms.Roles = append(perms.Roles, r2)
	}
	if err := roleRows.Err(); err != nil {
		return nil, err
	}

	userRows, err := r.db.Query(ctx, `
		SELECT cup.category_id, cup.user_id, u.display_name, u.avatar_url
		FROM category_user_permissions cup
		JOIN users u ON u.id = cup.user_id
		WHERE cup.category_id = $1
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()
	for userRows.Next() {
		u2 := domain.ChannelAllowedUser{}
		if err := userRows.Scan(&u2.ChannelID, &u2.UserID, &u2.DisplayName, &u2.AvatarURL); err != nil {
			return nil, err
		}
		u2.ChannelID = categoryID
		perms.Users = append(perms.Users, u2)
	}
	return perms, userRows.Err()
}

func (r *CategoryPermissionRepo) AddRole(ctx context.Context, categoryID, roleID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO category_role_permissions (category_id, role_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, categoryID, roleID)
	return err
}

func (r *CategoryPermissionRepo) RemoveRole(ctx context.Context, categoryID, roleID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM category_role_permissions WHERE category_id = $1 AND role_id = $2
	`, categoryID, roleID)
	return err
}

func (r *CategoryPermissionRepo) AddUser(ctx context.Context, categoryID, userID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO category_user_permissions (category_id, user_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, categoryID, userID)
	return err
}

func (r *CategoryPermissionRepo) RemoveUser(ctx context.Context, categoryID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM category_user_permissions WHERE category_id = $1 AND user_id = $2
	`, categoryID, userID)
	return err
}

func (r *CategoryPermissionRepo) HasAccess(ctx context.Context, categoryID, userID string, wsRoleIDs []string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM category_user_permissions
			WHERE category_id = $1 AND user_id = $2
			UNION ALL
			SELECT 1 FROM category_role_permissions
			WHERE category_id = $1 AND role_id = ANY($3)
		)
	`, categoryID, userID, wsRoleIDs).Scan(&exists)
	return exists, err
}

func (r *CategoryPermissionRepo) GetVisibleCategoryIDs(ctx context.Context, workspaceID, userID string, wsRoleIDs []string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id FROM channel_categories
		WHERE workspace_id = $1
		  AND (
		    is_private = FALSE
		    OR EXISTS (
		      SELECT 1 FROM category_user_permissions
		      WHERE category_id = channel_categories.id AND user_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM category_role_permissions
		      WHERE category_id = channel_categories.id AND role_id = ANY($3)
		    )
		  )
	`, workspaceID, userID, wsRoleIDs)
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
