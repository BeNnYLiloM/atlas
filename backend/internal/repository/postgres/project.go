package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/atlas/backend/internal/domain"
)

type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (id, workspace_id, name, description, icon_url, is_archived)
		VALUES ($1, $2, $3, $4, $5, FALSE)
		RETURNING created_at
	`
	return r.db.QueryRow(ctx, query,
		project.ID,
		project.WorkspaceID,
		project.Name,
		project.Description,
		project.IconURL,
	).Scan(&project.CreatedAt)
}

func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	query := `
		SELECT id, workspace_id, name, description, icon_url, is_archived, created_at
		FROM projects WHERE id = $1
	`
	p := &domain.Project{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.WorkspaceID, &p.Name, &p.Description,
		&p.IconURL, &p.IsArchived, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProjectRepo) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.Project, error) {
	query := `
		SELECT id, workspace_id, name, description, icon_url, is_archived, created_at
		FROM projects
		WHERE workspace_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.Name, &p.Description,
			&p.IconURL, &p.IsArchived, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Update(ctx context.Context, id string, update *domain.ProjectUpdate) (*domain.Project, error) {
	// icon_url нужно явно обнулять когда ClearIconURL=true, поэтому обрабатываем отдельно
	var iconURLExpr string
	if update.ClearIconURL {
		iconURLExpr = "icon_url = NULL"
	} else if update.IconURL != nil {
		iconURLExpr = "icon_url = $4"
	} else {
		iconURLExpr = "icon_url = icon_url"
	}

	query := `
		UPDATE projects
		SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			` + iconURLExpr + `
		WHERE id = $1
		RETURNING id, workspace_id, name, description, icon_url, is_archived, created_at
	`

	p := &domain.Project{}
	var err error
	if update.IconURL != nil && !update.ClearIconURL {
		err = r.db.QueryRow(ctx, query, id, update.Name, update.Description, update.IconURL).Scan(
			&p.ID, &p.WorkspaceID, &p.Name, &p.Description,
			&p.IconURL, &p.IsArchived, &p.CreatedAt,
		)
	} else {
		err = r.db.QueryRow(ctx, query, id, update.Name, update.Description).Scan(
			&p.ID, &p.WorkspaceID, &p.Name, &p.Description,
			&p.IconURL, &p.IsArchived, &p.CreatedAt,
		)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}

func (r *ProjectRepo) SetArchived(ctx context.Context, id string, archived bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET is_archived = $2 WHERE id = $1`,
		id, archived,
	)
	return err
}

func (r *ProjectRepo) AddMember(ctx context.Context, member *domain.ProjectMember) error {
	query := `
		INSERT INTO project_members (project_id, user_id, is_lead)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, user_id) DO NOTHING
		RETURNING joined_at
	`
	err := r.db.QueryRow(ctx, query, member.ProjectID, member.UserID, member.IsLead).
		Scan(&member.JoinedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func (r *ProjectRepo) RemoveMember(ctx context.Context, projectID, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`,
		projectID, userID,
	)
	return err
}

func (r *ProjectRepo) GetMembers(ctx context.Context, projectID string) ([]*domain.ProjectMember, error) {
	query := `
		SELECT pm.project_id, pm.user_id, pm.is_lead, pm.joined_at,
		       u.display_name, u.avatar_url
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY pm.is_lead DESC, pm.joined_at ASC
	`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	for rows.Next() {
		m := &domain.ProjectMember{}
		if err := rows.Scan(
			&m.ProjectID, &m.UserID, &m.IsLead, &m.JoinedAt,
			&m.DisplayName, &m.AvatarURL,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *ProjectRepo) GetMember(ctx context.Context, projectID, userID string) (*domain.ProjectMember, error) {
	query := `
		SELECT project_id, user_id, is_lead, joined_at
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	m := &domain.ProjectMember{}
	err := r.db.QueryRow(ctx, query, projectID, userID).Scan(
		&m.ProjectID, &m.UserID, &m.IsLead, &m.JoinedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *ProjectRepo) RemoveMemberFromAllProjects(ctx context.Context, workspaceID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM project_members
		WHERE user_id = $1
		  AND project_id IN (SELECT id FROM projects WHERE workspace_id = $2)
	`, userID, workspaceID)
	return err
}

func (r *ProjectRepo) SetLead(ctx context.Context, projectID, userID string, isLead bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE project_members SET is_lead = $3 WHERE project_id = $1 AND user_id = $2`,
		projectID, userID, isLead,
	)
	return err
}

func (r *ProjectRepo) GetLeadCount(ctx context.Context, projectID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM project_members WHERE project_id = $1 AND is_lead = TRUE`,
		projectID,
	).Scan(&count)
	return count, err
}
