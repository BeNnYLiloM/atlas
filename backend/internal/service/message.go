package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

type MessageService struct {
	messageRepo       repository.MessageRepository
	channelRepo       repository.ChannelRepository
	workspaceRepo     repository.WorkspaceRepository
	channelMemberRepo repository.ChannelMemberRepository
	permRepo          repository.ChannelPermissionRepository
	roleRepo          repository.WorkspaceRoleRepository
	projectRepo       repository.ProjectRepository
	dmRepo            repository.DMChannelRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	channelRepo repository.ChannelRepository,
	workspaceRepo repository.WorkspaceRepository,
	channelMemberRepo repository.ChannelMemberRepository,
	permRepo repository.ChannelPermissionRepository,
	roleRepo repository.WorkspaceRoleRepository,
	projectRepo repository.ProjectRepository,
	dmRepo repository.DMChannelRepository,
) *MessageService {
	return &MessageService{
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		workspaceRepo:     workspaceRepo,
		channelMemberRepo: channelMemberRepo,
		permRepo:          permRepo,
		roleRepo:          roleRepo,
		projectRepo:       projectRepo,
		dmRepo:            dmRepo,
	}
}

// Create создает новое сообщение
func (s *MessageService) Create(ctx context.Context, input domain.MessageCreate, userID string) (*domain.Message, error) {
	channel, member, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, input.ChannelID, userID)
	if err != nil {
		return nil, err
	}

	// Проверяем что проект не архивирован
	if channel.ProjectID != nil {
		project, err := s.projectRepo.GetByID(ctx, *channel.ProjectID)
		if err != nil {
			return nil, err
		}
		if project != nil && project.IsArchived {
			return nil, ErrProjectArchived
		}
	}

	// Slowmode: проверяем только для обычных сообщений (не тредов) и не для owner/admin.
	// member может быть nil для DM-каналов — в этом случае slowmode не применяется.
	memberRole := ""
	if member != nil {
		memberRole = member.Role
	}
	if channel.SlowmodeSeconds > 0 && input.ParentID == nil &&
		memberRole != domain.RoleOwner && memberRole != domain.RoleAdmin {
		lastAt, err := s.channelMemberRepo.GetLastMessageAt(ctx, userID, input.ChannelID)
		if err == nil && lastAt != nil {
			elapsed := time.Since(*lastAt)
			if elapsed < time.Duration(channel.SlowmodeSeconds)*time.Second {
				return nil, ErrSlowmode
			}
		}
	}

	message := &domain.Message{
		ID:        uuid.New().String(),
		ChannelID: input.ChannelID,
		UserID:    userID,
		Content:   input.Content,
		ParentID:  input.ParentID,
		Type:      domain.MessageTypeText,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Обновляем last_message_at для slowmode (только основные сообщения)
	if input.ParentID == nil && channel.SlowmodeSeconds > 0 {
		_ = s.channelMemberRepo.SetLastMessageAt(ctx, userID, input.ChannelID)
	}

	// Получаем сообщение с данными пользователя
	return s.messageRepo.GetByID(ctx, message.ID)
}

// GetByChannelID возвращает сообщения канала
func (s *MessageService) GetByChannelID(ctx context.Context, channelID, userID string, limit, offset int) ([]*domain.Message, error) {
	if _, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, channelID, userID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.messageRepo.GetByChannelID(ctx, channelID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Обогащаем сообщения thread unread counts
	for _, msg := range messages {
		if msg.ThreadRepliesCount > 0 {
			unreadCount, err := s.channelMemberRepo.GetThreadUnreadCount(ctx, userID, msg.ID)
			if err != nil {
				log.Printf("Error getting thread unread count for message %s: %v", msg.ID, err)
			} else {
				msg.ThreadUnreadCount = unreadCount
				log.Printf("Message %s: thread_replies=%d, thread_unread=%d", msg.ID, msg.ThreadRepliesCount, unreadCount)
			}
		}
	}

	return messages, nil
}

// GetThreadMessages возвращает сообщения треда
func (s *MessageService) GetThreadMessages(ctx context.Context, parentID, userID string) ([]*domain.Message, error) {
	// Получаем родительское сообщение
	parent, err := s.messageRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrMessageNotFound
	}

	if _, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, parent.ChannelID, userID); err != nil {
		return nil, err
	}

	return s.messageRepo.GetThreadMessages(ctx, parentID)
}

// Update обновляет сообщение
func (s *MessageService) Update(ctx context.Context, messageID string, input domain.MessageUpdate, userID string) (*domain.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	if _, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, message.ChannelID, userID); err != nil {
		return nil, err
	}

	// Только автор может редактировать сообщение
	if message.UserID != userID {
		return nil, ErrForbidden
	}

	message.Content = input.Content
	if err := s.messageRepo.Update(ctx, message); err != nil {
		return nil, err
	}

	return s.messageRepo.GetByID(ctx, messageID)
}

// Delete удаляет сообщение и возвращает channelID для WebSocket broadcast
func (s *MessageService) Delete(ctx context.Context, messageID, userID string) (string, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return "", err
	}
	if message == nil {
		return "", ErrMessageNotFound
	}

	channel, member, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, message.ChannelID, userID)
	if err != nil {
		return "", err
	}

	// Автор или admin/owner могут удалять.
	// Для DM member == nil — только автор может удалить своё сообщение.
	if message.UserID != userID {
		if member == nil || member.Role == domain.RoleMember {
			return "", ErrForbidden
		}
	}

	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		return "", err
	}

	return channel.ID, nil
}

// MarkThreadAsRead отмечает тред прочитанным
func (s *MessageService) MarkThreadAsRead(ctx context.Context, parentMessageID, userID string, lastMessageID *string) error {
	// Получаем родительское сообщение
	parent, err := s.messageRepo.GetByID(ctx, parentMessageID)
	if err != nil {
		return err
	}
	if parent == nil {
		return ErrMessageNotFound
	}

	if _, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, parent.ChannelID, userID); err != nil {
		return err
	}

	return s.channelMemberRepo.MarkThreadAsRead(ctx, userID, parentMessageID, lastMessageID)
}

// GetThreadUnreadCount возвращает количество непрочитанных сообщений в треде
func (s *MessageService) GetThreadUnreadCount(ctx context.Context, parentMessageID, userID string) (int, error) {
	// Получаем родительское сообщение
	parent, err := s.messageRepo.GetByID(ctx, parentMessageID)
	if err != nil {
		return 0, err
	}
	if parent == nil {
		return 0, ErrMessageNotFound
	}

	if _, _, err := getAccessibleChannel(ctx, s.channelRepo, s.workspaceRepo, s.roleRepo, s.permRepo, s.projectRepo, s.dmRepo, parent.ChannelID, userID); err != nil {
		return 0, err
	}

	return s.channelMemberRepo.GetThreadUnreadCount(ctx, userID, parentMessageID)
}

// CreateCallMessage создаёт системное сообщение о звонке в DM-канале.
// Начальный статус = ringing (ждём ответа); меняется на missed/ongoing/ended через UpdateCallStatus.
func (s *MessageService) CreateCallMessage(ctx context.Context, channelID, callerID string) (*domain.Message, error) {
	status := domain.CallStatusRinging
	msg := &domain.Message{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		UserID:      callerID,
		Content:     "",
		Type:        domain.MessageTypeCall,
		CallStatus:  &status,
	}
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("CreateCallMessage: %w", err)
	}
	return s.messageRepo.GetByID(ctx, msg.ID)
}

// AcceptCallStatus вызывается получателем звонка (signal=accepted).
// Проверяет что сообщение принадлежит данному каналу, переводит ringing→ongoing.
// owner-check не применяется т.к. обновляет получатель, а не создатель.
func (s *MessageService) AcceptCallStatus(ctx context.Context, messageID, channelID string) error {
	// Безопасность: проверяем что сообщение принадлежит указанному каналу
	msg, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("AcceptCallStatus: %w", err)
	}
	if msg == nil || msg.ChannelID != channelID || msg.Type != domain.MessageTypeCall {
		return fmt.Errorf("AcceptCallStatus: message not found in channel")
	}
	// Пустой callerID — owner-check пропускается в SQL, переход ringing→ongoing
	if err := s.messageRepo.UpdateCallStatus(ctx, messageID, msg.UserID, domain.CallStatusOngoing, nil,
		[]string{domain.CallStatusRinging}); err != nil {
		return fmt.Errorf("AcceptCallStatus: %w", err)
	}
	return nil
}

// EndCallStatus вызывается любым участником DM (signal=ended/cancelled/missed).
// Проверяет принадлежность канала, разрешает переходы ringing→missed/cancelled, ongoing→ended.
func (s *MessageService) EndCallStatus(ctx context.Context, messageID, channelID, finalStatus string, durationSec *int) error {
	allowedFinalStatuses := map[string]bool{
		domain.CallStatusMissed:    true,
		domain.CallStatusCancelled: true,
		domain.CallStatusEnded:     true,
	}
	if !allowedFinalStatuses[finalStatus] {
		return fmt.Errorf("EndCallStatus: invalid final status %q", finalStatus)
	}

	msg, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("EndCallStatus: %w", err)
	}
	if msg == nil || msg.ChannelID != channelID || msg.Type != domain.MessageTypeCall {
		return fmt.Errorf("EndCallStatus: message not found in channel")
	}

	// Определяем допустимые исходные статусы для данного перехода
	var allowedFrom []string
	switch finalStatus {
	case domain.CallStatusEnded:
		allowedFrom = []string{domain.CallStatusOngoing}
	case domain.CallStatusMissed, domain.CallStatusCancelled:
		allowedFrom = []string{domain.CallStatusRinging}
	}

	if err := s.messageRepo.UpdateCallStatus(ctx, messageID, msg.UserID, finalStatus, durationSec, allowedFrom); err != nil {
		return fmt.Errorf("EndCallStatus: %w", err)
	}
	return nil
}
