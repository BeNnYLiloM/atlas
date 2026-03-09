package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type ChannelPermissionRepo struct {
	db *pgxpool.Pool
}

func NewChannelPermissionRepo(db *pgxpool.Pool) *ChannelPermissionRepo {
	return &ChannelPermissionRepo{db: db}
}

func (r *ChannelPermissionRepo) GetPermissions(ctx context.Context, channelID string) (*domain.ChannelPermissions, error) {
	perms := &domain.ChannelPermissions{
		Roles: []domain.ChannelAllowedRole{},
		Users: []domain.ChannelAllowedUser{},
	}

	// Роли с доступом
	roleRows, err := r.db.Query(ctx, `
		SELECT crp.channel_id, crp.role_id, wr.name, wr.color
		FROM channel_role_permissions crp
		JOIN workspace_roles wr ON wr.id = crp.role_id
		WHERE crp.channel_id = $1
		ORDER BY wr.position DESC
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()

	for roleRows.Next() {
		r2 := domain.ChannelAllowedRole{}
		if err := roleRows.Scan(&r2.ChannelID, &r2.RoleID, &r2.RoleName, &r2.RoleColor); err != nil {
			return nil, err
		}
		perms.Roles = append(perms.Roles, r2)
	}
	if err := roleRows.Err(); err != nil {
		return nil, err
	}

	// Участники с доступом
	userRows, err := r.db.Query(ctx, `
		SELECT cup.channel_id, cup.user_id, u.display_name, u.avatar_url
		FROM channel_user_permissions cup
		JOIN users u ON u.id = cup.user_id
		WHERE cup.channel_id = $1
		ORDER BY u.display_name
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()

	for userRows.Next() {
		u := domain.ChannelAllowedUser{}
		if err := userRows.Scan(&u.ChannelID, &u.UserID, &u.DisplayName, &u.AvatarURL); err != nil {
			return nil, err
		}
		perms.Users = append(perms.Users, u)
	}

	return perms, userRows.Err()
}

func (r *ChannelPermissionRepo) AddRole(ctx context.Context, channelID, roleID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_role_permissions (channel_id, role_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, channelID, roleID)
	return err
}

func (r *ChannelPermissionRepo) RemoveRole(ctx context.Context, channelID, roleID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM channel_role_permissions WHERE channel_id = $1 AND role_id = $2
	`, channelID, roleID)
	return err
}

func (r *ChannelPermissionRepo) AddUser(ctx context.Context, channelID, userID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO channel_user_permissions (channel_id, user_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, channelID, userID)
	return err
}

func (r *ChannelPermissionRepo) RemoveUser(ctx context.Context, channelID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM channel_user_permissions WHERE channel_id = $1 AND user_id = $2
	`, channelID, userID)
	return err
}

// GetChannelsByRole — все каналы где данная роль добавлена в channel_role_permissions
func (r *ChannelPermissionRepo) GetChannelsByRole(ctx context.Context, roleID string) ([]*domain.Channel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.workspace_id, c.name, c.type, c.is_private,
		       c.topic, c.slowmode_seconds, c.position, c.category_id, c.created_at
		FROM channels c
		JOIN channel_role_permissions crp ON crp.channel_id = c.id
		WHERE crp.role_id = $1
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*domain.Channel
	for rows.Next() {
		ch := &domain.Channel{}
		if err := rows.Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Type, &ch.IsPrivate,
			&ch.Topic, &ch.SlowmodeSeconds, &ch.Position, &ch.CategoryID, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

// HasAccess — есть ли у пользователя явный доступ к приватному каналу
// (личный доступ ИЛИ через одну из его ролей)
func (r *ChannelPermissionRepo) HasAccess(ctx context.Context, channelID, userID string, wsRoleIDs []string) (bool, error) {
	// Личный доступ
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM channel_user_permissions
			WHERE channel_id = $1 AND user_id = $2
		)
	`, channelID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// Доступ через роль
	if len(wsRoleIDs) == 0 {
		return false, nil
	}
	err = r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM channel_role_permissions
			WHERE channel_id = $1 AND role_id = ANY($2)
		)
	`, channelID, wsRoleIDs).Scan(&exists)
	return exists, err
}
