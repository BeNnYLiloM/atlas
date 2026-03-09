package service

import (
	"context"
	"time"

	"github.com/your-org/atlas/backend/internal/repository/postgres"
)

type SearchService struct {
	repo *postgres.SearchRepository
}

func NewSearchService(repo *postgres.SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

type SearchParams struct {
	Query       string
	WorkspaceID string
	ChannelID   string
	UserID      string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

type SearchResponse struct {
	Results []*postgres.SearchResult `json:"results"`
	Total   int                      `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
}

func (s *SearchService) Search(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	if params.Query == "" {
		return &SearchResponse{Results: []*postgres.SearchResult{}, Total: 0}, nil
	}
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 20
	}

	results, total, err := s.repo.Search(ctx, postgres.SearchFilter{
		Query:       params.Query,
		WorkspaceID: params.WorkspaceID,
		ChannelID:   params.ChannelID,
		UserID:      params.UserID,
		From:        params.From,
		To:          params.To,
		Limit:       params.Limit,
		Offset:      params.Offset,
	})
	if err != nil {
		return nil, err
	}

	if results == nil {
		results = []*postgres.SearchResult{}
	}

	return &SearchResponse{
		Results: results,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}, nil
}
