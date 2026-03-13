package service

import (
	"context"
	"time"

	"github.com/your-org/atlas/backend/internal/repository"
)

type SearchService struct {
	repo          repository.SearchRepository
	workspaceRepo repository.WorkspaceRepository
	channelRepo   repository.ChannelRepository
	roleRepo      repository.WorkspaceRoleRepository
	permRepo      repository.ChannelPermissionRepository
}

func NewSearchService(
	repo repository.SearchRepository,
	workspaceRepo repository.WorkspaceRepository,
	channelRepo repository.ChannelRepository,
	roleRepo repository.WorkspaceRoleRepository,
	permRepo repository.ChannelPermissionRepository,
) *SearchService {
	return &SearchService{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		channelRepo:   channelRepo,
		roleRepo:      roleRepo,
		permRepo:      permRepo,
	}
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
	Results []*repository.SearchResult `json:"results"`
	Total   int                        `json:"total"`
	Limit   int                        `json:"limit"`
	Offset  int                        `json:"offset"`
}

func (s *SearchService) Search(ctx context.Context, actorUserID string, params SearchParams) (*SearchResponse, error) {
	if params.Query == "" {
		return &SearchResponse{Results: []*repository.SearchResult{}, Total: 0}, nil
	}
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 20
	}

	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, params.WorkspaceID, actorUserID); err != nil {
		return nil, err
	}

	if params.ChannelID != "" {
		channel, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, params.ChannelID, actorUserID)
		if err != nil {
			return nil, err
		}
		if channel.WorkspaceID != params.WorkspaceID {
			return nil, ErrForbidden
		}
	}

	results, total, err := s.repo.Search(ctx, repository.SearchFilter{
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
		results = []*repository.SearchResult{}
	}

	return &SearchResponse{
		Results: results,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}, nil
}
