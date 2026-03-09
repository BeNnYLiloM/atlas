package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type SearchResult struct {
	Message   *domain.Message `json:"message"`
	Rank      float32         `json:"rank"`
	Highlight string          `json:"highlight"`
}

type SearchFilter struct {
	Query       string
	WorkspaceID string
	ChannelID   string
	UserID      string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

type SearchRepository struct {
	db *pgxpool.Pool
}

func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) Search(ctx context.Context, filter SearchFilter) ([]*SearchResult, int, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	args := []interface{}{}
	conditions := []string{}
	argIdx := 1

	// Поисковый запрос (обязательный)
	args = append(args, filter.Query)
	tsQuery := fmt.Sprintf("plainto_tsquery('russian', $%d)", argIdx)
	argIdx++

	conditions = append(conditions,
		fmt.Sprintf("m.search_vector @@ %s", tsQuery))

	// Фильтр по workspace через каналы
	if filter.WorkspaceID != "" {
		args = append(args, filter.WorkspaceID)
		conditions = append(conditions, fmt.Sprintf("c.workspace_id = $%d", argIdx))
		argIdx++
	}

	// Фильтр по каналу
	if filter.ChannelID != "" {
		args = append(args, filter.ChannelID)
		conditions = append(conditions, fmt.Sprintf("m.channel_id = $%d", argIdx))
		argIdx++
	}

	// Фильтр по автору
	if filter.UserID != "" {
		args = append(args, filter.UserID)
		conditions = append(conditions, fmt.Sprintf("m.user_id = $%d", argIdx))
		argIdx++
	}

	// Фильтр по дате
	if filter.From != nil {
		args = append(args, filter.From)
		conditions = append(conditions, fmt.Sprintf("m.created_at >= $%d", argIdx))
		argIdx++
	}
	if filter.To != nil {
		args = append(args, filter.To)
		conditions = append(conditions, fmt.Sprintf("m.created_at <= $%d", argIdx))
		argIdx++
	}

	// Исключаем ответы в тредах из основного поиска
	conditions = append(conditions, "m.parent_id IS NULL")

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM messages m
		JOIN channels c ON m.channel_id = c.id
		WHERE %s
	`, whereClause)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("SearchRepository.Search count: %w", err)
	}

	args = append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`
		SELECT 
			m.id, m.channel_id, m.user_id, m.content, m.parent_id, m.created_at, m.updated_at,
			u.id, u.display_name, u.avatar_url,
			ts_rank(m.search_vector, %s) AS rank,
			ts_headline('russian', m.content, %s, 'MaxWords=15, MinWords=5') AS highlight
		FROM messages m
		JOIN channels c ON m.channel_id = c.id
		LEFT JOIN users u ON m.user_id = u.id
		WHERE %s
		ORDER BY rank DESC, m.created_at DESC
		LIMIT $%d OFFSET $%d
	`, tsQuery, tsQuery, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("SearchRepository.Search query: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		msg := &domain.Message{}
		user := &domain.User{}
		result := &SearchResult{Message: msg}

		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.ParentID, &msg.CreatedAt, &msg.UpdatedAt,
			&user.ID, &user.DisplayName, &user.AvatarURL,
			&result.Rank,
			&result.Highlight,
		); err != nil {
			return nil, 0, fmt.Errorf("SearchRepository.Search scan: %w", err)
		}
		msg.User = user
		results = append(results, result)
	}

	return results, total, nil
}
