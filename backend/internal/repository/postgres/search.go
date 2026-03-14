package postgres

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

var nonWordRe = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// buildPrefixTsQuery преобразует произвольный ввод в prefix tsquery:
// "прив мир" → "прив:* & мир:*"
func buildPrefixTsQuery(q string) string {
	normalized := nonWordRe.ReplaceAllString(strings.TrimSpace(q), " ")
	words := strings.Fields(normalized)
	if len(words) == 0 {
		return ""
	}
	tokens := make([]string, 0, len(words))
	for _, w := range words {
		tokens = append(tokens, w+":*")
	}
	return strings.Join(tokens, " & ")
}


type SearchRepository struct {
	db *pgxpool.Pool
}

func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{db: db}
}

var _ repository.SearchRepository = (*SearchRepository)(nil)

func (r *SearchRepository) Search(ctx context.Context, filter repository.SearchFilter) ([]*repository.SearchResult, int, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	args := []interface{}{}
	conditions := []string{}
	argIdx := 1

	// Поисковый запрос (обязательный): prefix search + ILIKE fallback
	prefixQuery := buildPrefixTsQuery(filter.Query)
	var tsQuery string
	if prefixQuery != "" {
		args = append(args, prefixQuery)
		tsQuery = fmt.Sprintf("to_tsquery('russian', $%d)", argIdx)
		argIdx++
		// ILIKE использует GIN trigram-индекс (idx_messages_content_trgm) — не seq scan
		args = append(args, filter.Query)
		conditions = append(conditions,
			fmt.Sprintf("(m.search_vector @@ %s OR m.content ILIKE '%%' || $%d || '%%')", tsQuery, argIdx))
		argIdx++
	} else {
		// Ввод из одних спецсимволов — не ищем
		return nil, 0, nil
	}

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

	var rankExpr, headlineExpr string
	if prefixQuery != "" {
		rankExpr = fmt.Sprintf("ts_rank(m.search_vector, %s)", tsQuery)
		headlineExpr = fmt.Sprintf("ts_headline('russian', m.content, %s, 'MaxWords=15, MinWords=5')", tsQuery)
	} else {
		rankExpr = "0.0::float4"
		headlineExpr = "m.content"
	}

	query := fmt.Sprintf(`
		SELECT 
			m.id, m.channel_id, m.user_id, m.content, m.parent_id, m.created_at, m.updated_at,
			u.id, u.display_name, u.avatar_url,
			%s AS rank,
			%s AS highlight
		FROM messages m
		JOIN channels c ON m.channel_id = c.id
		LEFT JOIN users u ON m.user_id = u.id
		WHERE %s
		ORDER BY rank DESC, m.created_at DESC
		LIMIT $%d OFFSET $%d
	`, rankExpr, headlineExpr, whereClause, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("SearchRepository.Search query: %w", err)
	}
	defer rows.Close()

	var results []*repository.SearchResult
	for rows.Next() {
		msg := &domain.Message{}
		author := &domain.MessageAuthor{}
		result := &repository.SearchResult{Message: msg}

		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.ParentID, &msg.CreatedAt, &msg.UpdatedAt,
			&author.ID, &author.DisplayName, &author.AvatarURL,
			&result.Rank,
			&result.Highlight,
		); err != nil {
			return nil, 0, fmt.Errorf("SearchRepository.Search scan: %w", err)
		}
		msg.User = author
		results = append(results, result)
	}

	return results, total, nil
}
