package service

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/atlas/backend/internal/domain"
)

type mockWorkspaceRepo struct {
	members map[string]map[string]*domain.WorkspaceMember
}

func newMockWorkspaceRepo() *mockWorkspaceRepo {
	return &mockWorkspaceRepo{members: map[string]map[string]*domain.WorkspaceMember{}}
}

func (m *mockWorkspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	return nil
}
func (m *mockWorkspaceRepo) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	return nil, nil
}
func (m *mockWorkspaceRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	return nil, nil
}
func (m *mockWorkspaceRepo) Update(ctx context.Context, id string, update *domain.WorkspaceUpdate) (*domain.Workspace, error) {
	return nil, nil
}
func (m *mockWorkspaceRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockWorkspaceRepo) AddMember(ctx context.Context, member *domain.WorkspaceMember) error {
	if m.members[member.WorkspaceID] == nil {
		m.members[member.WorkspaceID] = map[string]*domain.WorkspaceMember{}
	}
	m.members[member.WorkspaceID][member.UserID] = member
	return nil
}
func (m *mockWorkspaceRepo) GetMembers(ctx context.Context, workspaceID string) ([]*domain.WorkspaceMember, error) {
	var result []*domain.WorkspaceMember
	for _, member := range m.members[workspaceID] {
		result = append(result, member)
	}
	return result, nil
}
func (m *mockWorkspaceRepo) GetMemberUserIDs(ctx context.Context, workspaceID string) ([]string, error) {
	var ids []string
	for userID := range m.members[workspaceID] {
		ids = append(ids, userID)
	}
	return ids, nil
}
func (m *mockWorkspaceRepo) GetMember(ctx context.Context, workspaceID, userID string) (*domain.WorkspaceMember, error) {
	if m.members[workspaceID] == nil {
		return nil, nil
	}
	return m.members[workspaceID][userID], nil
}
func (m *mockWorkspaceRepo) UpdateMember(ctx context.Context, workspaceID, userID string, update *domain.WorkspaceMemberUpdate) error {
	return nil
}
func (m *mockWorkspaceRepo) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	return nil
}

type mockChannelRepo struct {
	channels map[string]*domain.Channel
}

func (m *mockChannelRepo) Create(ctx context.Context, channel *domain.Channel) error { return nil }
func (m *mockChannelRepo) GetByID(ctx context.Context, id string) (*domain.Channel, error) {
	return m.channels[id], nil
}
func (m *mockChannelRepo) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.Channel, error) {
	var result []*domain.Channel
	for _, channel := range m.channels {
		if channel.WorkspaceID == workspaceID {
			result = append(result, channel)
		}
	}
	return result, nil
}
func (m *mockChannelRepo) GetVisibleByWorkspaceID(ctx context.Context, workspaceID, userID string, roleIDs []string) ([]*domain.Channel, error) {
	return m.GetByWorkspaceID(ctx, workspaceID)
}
func (m *mockChannelRepo) Update(ctx context.Context, id string, update *domain.ChannelUpdate) (*domain.Channel, error) {
	return m.channels[id], nil
}
func (m *mockChannelRepo) Delete(ctx context.Context, id string) error { return nil }

type mockRoleRepo struct {
	memberRoles map[string]map[string][]*domain.WorkspaceRole
}

func newMockRoleRepo() *mockRoleRepo {
	return &mockRoleRepo{memberRoles: map[string]map[string][]*domain.WorkspaceRole{}}
}

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.WorkspaceRole) error { return nil }
func (m *mockRoleRepo) GetByID(ctx context.Context, id string) (*domain.WorkspaceRole, error) {
	return nil, nil
}
func (m *mockRoleRepo) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.WorkspaceRole, error) {
	return nil, nil
}
func (m *mockRoleRepo) Update(ctx context.Context, id string, update *domain.WorkspaceRoleUpdate) (*domain.WorkspaceRole, error) {
	return nil, nil
}
func (m *mockRoleRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockRoleRepo) AssignRole(ctx context.Context, workspaceID, userID, roleID string) error {
	return nil
}
func (m *mockRoleRepo) RevokeRole(ctx context.Context, workspaceID, userID, roleID string) error {
	return nil
}
func (m *mockRoleRepo) GetMemberRoles(ctx context.Context, workspaceID, userID string) ([]*domain.WorkspaceRole, error) {
	if m.memberRoles[workspaceID] == nil {
		return nil, nil
	}
	return m.memberRoles[workspaceID][userID], nil
}
func (m *mockRoleRepo) GetUserIDsByRole(ctx context.Context, roleID string) ([]string, error) {
	return nil, nil
}
func (m *mockRoleRepo) GetEffectivePermissions(ctx context.Context, workspaceID, userID string) (*domain.RolePermissions, error) {
	return nil, nil
}

type mockChannelPermissionRepo struct {
	allowed map[string]map[string]bool
}

func newMockChannelPermissionRepo() *mockChannelPermissionRepo {
	return &mockChannelPermissionRepo{allowed: map[string]map[string]bool{}}
}

func (m *mockChannelPermissionRepo) GetPermissions(ctx context.Context, channelID string) (*domain.ChannelPermissions, error) {
	return &domain.ChannelPermissions{}, nil
}
func (m *mockChannelPermissionRepo) AddRole(ctx context.Context, channelID, roleID string) error {
	return nil
}
func (m *mockChannelPermissionRepo) RemoveRole(ctx context.Context, channelID, roleID string) error {
	return nil
}
func (m *mockChannelPermissionRepo) AddUser(ctx context.Context, channelID, userID string) error {
	return nil
}
func (m *mockChannelPermissionRepo) RemoveUser(ctx context.Context, channelID, userID string) error {
	return nil
}
func (m *mockChannelPermissionRepo) HasAccess(ctx context.Context, channelID, userID string, wsRoleIDs []string) (bool, error) {
	if m.allowed[channelID] == nil {
		return false, nil
	}
	return m.allowed[channelID][userID], nil
}
func (m *mockChannelPermissionRepo) GetChannelsByRole(ctx context.Context, roleID string) ([]*domain.Channel, error) {
	return nil, nil
}

type mockMessageRepo struct {
	messages          map[string]*domain.Message
	messagesByChannel map[string][]*domain.Message
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages:          map[string]*domain.Message{},
		messagesByChannel: map[string][]*domain.Message{},
	}
}

func (m *mockMessageRepo) Create(ctx context.Context, message *domain.Message) error {
	m.messages[message.ID] = message
	m.messagesByChannel[message.ChannelID] = append(m.messagesByChannel[message.ChannelID], message)
	return nil
}
func (m *mockMessageRepo) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	return m.messages[id], nil
}
func (m *mockMessageRepo) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	return m.messagesByChannel[channelID], nil
}
func (m *mockMessageRepo) GetThreadMessages(ctx context.Context, parentID string) ([]*domain.Message, error) {
	return nil, nil
}
func (m *mockMessageRepo) Update(ctx context.Context, message *domain.Message) error { return nil }
func (m *mockMessageRepo) Delete(ctx context.Context, id string) error               { return nil }

type mockChannelMemberRepo struct{}

func (m *mockChannelMemberRepo) UpsertMember(ctx context.Context, userID, channelID string) error {
	return nil
}
func (m *mockChannelMemberRepo) RemoveMember(ctx context.Context, userID, channelID string) error {
	return nil
}
func (m *mockChannelMemberRepo) GetMembers(ctx context.Context, channelID string) ([]*domain.ChannelMemberInfo, error) {
	return nil, nil
}
func (m *mockChannelMemberRepo) MarkAsRead(ctx context.Context, userID, channelID string, messageID *string) error {
	return nil
}
func (m *mockChannelMemberRepo) GetUnreadCount(ctx context.Context, userID, channelID string) (int, error) {
	return 0, nil
}
func (m *mockChannelMemberRepo) GetUnreadCountsForWorkspace(ctx context.Context, userID, workspaceID string) (map[string]domain.ChannelStats, error) {
	return map[string]domain.ChannelStats{}, nil
}
func (m *mockChannelMemberRepo) GetLastReadMessageID(ctx context.Context, userID, channelID string) (*string, error) {
	return nil, nil
}
func (m *mockChannelMemberRepo) UpdateNotificationLevel(ctx context.Context, userID, channelID, level string) error {
	return nil
}
func (m *mockChannelMemberRepo) GetNotificationLevel(ctx context.Context, userID, channelID string) (string, error) {
	return "", nil
}
func (m *mockChannelMemberRepo) GetLastMessageAt(ctx context.Context, userID, channelID string) (*time.Time, error) {
	return nil, nil
}
func (m *mockChannelMemberRepo) SetLastMessageAt(ctx context.Context, userID, channelID string) error {
	return nil
}
func (m *mockChannelMemberRepo) MarkThreadAsRead(ctx context.Context, userID, parentMessageID string, lastMessageID *string) error {
	return nil
}
func (m *mockChannelMemberRepo) GetThreadUnreadCount(ctx context.Context, userID, parentMessageID string) (int, error) {
	return 0, nil
}

func TestMessageService_GetByChannelID_DeniesPrivateChannelWithoutPermission(t *testing.T) {
	workspaceRepo := newMockWorkspaceRepo()
	workspaceRepo.AddMember(context.Background(), &domain.WorkspaceMember{WorkspaceID: "ws-1", UserID: "user-1", Role: domain.RoleMember})

	messageService := NewMessageService(
		newMockMessageRepo(),
		&mockChannelRepo{channels: map[string]*domain.Channel{
			"ch-1": {ID: "ch-1", WorkspaceID: "ws-1", IsPrivate: true},
		}},
		workspaceRepo,
		&mockChannelMemberRepo{},
		newMockChannelPermissionRepo(),
		newMockRoleRepo(),
	)

	_, err := messageService.GetByChannelID(context.Background(), "ch-1", "user-1", 50, 0)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMessageService_GetByChannelID_AllowsPrivateChannelWithPermission(t *testing.T) {
	workspaceRepo := newMockWorkspaceRepo()
	workspaceRepo.AddMember(context.Background(), &domain.WorkspaceMember{WorkspaceID: "ws-1", UserID: "user-1", Role: domain.RoleMember})

	messageRepo := newMockMessageRepo()
	messageRepo.messagesByChannel["ch-1"] = []*domain.Message{{ID: "msg-1", ChannelID: "ch-1", UserID: "user-2", Content: "hello"}}

	permRepo := newMockChannelPermissionRepo()
	permRepo.allowed["ch-1"] = map[string]bool{"user-1": true}

	messageService := NewMessageService(
		messageRepo,
		&mockChannelRepo{channels: map[string]*domain.Channel{
			"ch-1": {ID: "ch-1", WorkspaceID: "ws-1", IsPrivate: true},
		}},
		workspaceRepo,
		&mockChannelMemberRepo{},
		permRepo,
		newMockRoleRepo(),
	)

	messages, err := messageService.GetByChannelID(context.Background(), "ch-1", "user-1", 50, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(messages) != 1 || messages[0].ID != "msg-1" {
		t.Fatalf("expected allowed access to channel messages, got %#v", messages)
	}
}

func TestSearchService_Search_DeniesNonMember(t *testing.T) {
	searchService := NewSearchService(nil, newMockWorkspaceRepo(), &mockChannelRepo{}, newMockRoleRepo(), newMockChannelPermissionRepo())

	_, err := searchService.Search(context.Background(), "user-1", SearchParams{
		Query:       "hello",
		WorkspaceID: "ws-1",
	})
	if err != ErrNotMember {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestSearchService_Search_DeniesPrivateChannelWithoutPermission(t *testing.T) {
	workspaceRepo := newMockWorkspaceRepo()
	workspaceRepo.AddMember(context.Background(), &domain.WorkspaceMember{WorkspaceID: "ws-1", UserID: "user-1", Role: domain.RoleMember})

	searchService := NewSearchService(
		nil,
		workspaceRepo,
		&mockChannelRepo{channels: map[string]*domain.Channel{
			"ch-1": {ID: "ch-1", WorkspaceID: "ws-1", IsPrivate: true},
		}},
		newMockRoleRepo(),
		newMockChannelPermissionRepo(),
	)

	_, err := searchService.Search(context.Background(), "user-1", SearchParams{
		Query:       "hello",
		WorkspaceID: "ws-1",
		ChannelID:   "ch-1",
	})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_Update_DeniesRegularMemberWhoDoesNotOwnTask(t *testing.T) {
	workspaceRepo := newMockWorkspaceRepo()
	workspaceRepo.AddMember(context.Background(), &domain.WorkspaceMember{WorkspaceID: "ws-1", UserID: "user-1", Role: domain.RoleMember})

	repo := newMockTaskRepository(map[string]*domain.Task{
		"task-1": {ID: "task-1", WorkspaceID: "ws-1", ReporterID: strPtr("reporter-1")},
	})

	taskService := NewTaskService(repo, workspaceRepo, newMockMessageRepo(), &mockChannelRepo{}, newMockRoleRepo(), newMockChannelPermissionRepo())

	update := domain.TaskUpdate{Title: strPtr("new title")}
	err := taskService.Update(context.Background(), "task-1", "user-1", &update)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_Delete_AllowsWorkspaceAdmin(t *testing.T) {
	workspaceRepo := newMockWorkspaceRepo()
	workspaceRepo.AddMember(context.Background(), &domain.WorkspaceMember{WorkspaceID: "ws-1", UserID: "admin-1", Role: domain.RoleAdmin})

	repo := newMockTaskRepository(map[string]*domain.Task{
		"task-1": {ID: "task-1", WorkspaceID: "ws-1", ReporterID: strPtr("reporter-1")},
	})

	taskService := NewTaskService(repo, workspaceRepo, newMockMessageRepo(), &mockChannelRepo{}, newMockRoleRepo(), newMockChannelPermissionRepo())

	if err := taskService.Delete(context.Background(), "task-1", "admin-1"); err != nil {
		t.Fatalf("expected admin delete to succeed, got %v", err)
	}
	if !repo.deleted["task-1"] {
		t.Fatal("expected task repository delete to be called")
	}
}

type mockTaskRepository struct {
	tasks   map[string]*domain.Task
	deleted map[string]bool
}

func newMockTaskRepository(tasks map[string]*domain.Task) *mockTaskRepository {
	return &mockTaskRepository{tasks: tasks, deleted: map[string]bool{}}
}

func (m *mockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	m.tasks[task.ID] = task
	return nil
}
func (m *mockTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return m.tasks[id], nil
}
func (m *mockTaskRepository) GetByWorkspace(ctx context.Context, workspaceID string, status string) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, task := range m.tasks {
		if task.WorkspaceID == workspaceID {
			result = append(result, task)
		}
	}
	return result, nil
}
func (m *mockTaskRepository) Update(ctx context.Context, id string, update *domain.TaskUpdate) error {
	return nil
}
func (m *mockTaskRepository) Delete(ctx context.Context, id string) error {
	m.deleted[id] = true
	return nil
}

func strPtr(value string) *string { return &value }
