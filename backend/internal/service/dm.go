package service

import (
	"context"
	"fmt"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

type DMService struct {
	dmRepo        repository.DMChannelRepository
	userRepo      repository.UserRepository
	workspaceRepo repository.WorkspaceRepository
}

func NewDMService(
	dmRepo repository.DMChannelRepository,
	userRepo repository.UserRepository,
	workspaceRepo repository.WorkspaceRepository,
) *DMService {
	return &DMService{
		dmRepo:        dmRepo,
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
	}
}

// GetOrCreateDM открывает или возвращает существующий DM-канал между двумя участниками.
// Валидации:
//   - нельзя создать DM с самим собой
//   - target user должен существовать и быть членом воркспейса
func (s *DMService) GetOrCreateDM(ctx context.Context, workspaceID, initiatorID, targetUserID string) (*domain.Channel, error) {
	if initiatorID == targetUserID {
		return nil, ErrDMSelf
	}

	// Инициатор должен быть членом воркспейса
	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, workspaceID, initiatorID); err != nil {
		return nil, err
	}

	// Target должен существовать и быть членом воркспейса
	targetMember, err := s.workspaceRepo.GetMember(ctx, workspaceID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("dm: get target member: %w", err)
	}
	if targetMember == nil {
		return nil, ErrNotMember
	}

	return s.dmRepo.GetOrCreate(ctx, workspaceID, initiatorID, targetUserID)
}

// ListDMs возвращает все DM-диалоги пользователя в воркспейсе.
func (s *DMService) ListDMs(ctx context.Context, workspaceID, userID string) ([]*domain.DMChannel, error) {
	if _, err := ensureWorkspaceMember(ctx, s.workspaceRepo, workspaceID, userID); err != nil {
		return nil, err
	}

	return s.dmRepo.GetByUserID(ctx, workspaceID, userID)
}
