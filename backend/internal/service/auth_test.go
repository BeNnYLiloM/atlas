package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/atlas/backend/internal/config"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	return nil
}

type mockAuthSessionRepo struct {
	sessions map[string]*domain.AuthSession
}

func newMockAuthSessionRepo() *mockAuthSessionRepo {
	return &mockAuthSessionRepo{sessions: make(map[string]*domain.AuthSession)}
}

func (m *mockAuthSessionRepo) Create(ctx context.Context, session *domain.AuthSession) error {
	clone := *session
	m.sessions[session.RefreshTokenHash] = &clone
	return nil
}

func (m *mockAuthSessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.AuthSession, error) {
	session, ok := m.sessions[hash]
	if !ok {
		return nil, nil
	}
	clone := *session
	return &clone, nil
}

func (m *mockAuthSessionRepo) Rotate(ctx context.Context, currentSessionID string, nextSession *domain.AuthSession) error {
	for _, session := range m.sessions {
		if session.ID == currentSessionID {
			now := nextSession.CreatedAt
			session.RevokedAt = &now
			replacedBy := nextSession.ID
			session.ReplacedBySessionID = &replacedBy
			clone := *nextSession
			m.sessions[nextSession.RefreshTokenHash] = &clone
			return nil
		}
	}
	return nil
}

func (m *mockAuthSessionRepo) RevokeByID(ctx context.Context, sessionID string) error {
	for _, session := range m.sessions {
		if session.ID == sessionID && session.RevokedAt == nil {
			now := time.Now().UTC()
			session.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockAuthSessionRepo) RevokeFamily(ctx context.Context, familyID string) error {
	now := time.Now().UTC()
	for _, session := range m.sessions {
		if session.FamilyID == familyID && session.RevokedAt == nil {
			session.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockAuthSessionRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	for _, session := range m.sessions {
		if session.UserID == userID && session.RevokedAt == nil {
			session.RevokedAt = &now
		}
	}
	return nil
}

func newTestAuthService() *service.AuthService {
	return newTestAuthServiceWithRepos(newMockUserRepo(), newMockAuthSessionRepo())
}

func newTestAuthServiceWithRepos(userRepo *mockUserRepo, sessionRepo *mockAuthSessionRepo) *service.AuthService {
	cfg := config.JWTConfig{
		Secret:                "test-secret-key",
		AccessTokenTTLMinutes: 15,
		RefreshTokenTTLDays:   7,
		Issuer:                "atlas-test",
		Audience:              "atlas-web",
		RefreshCookieName:     "atlas_refresh_token",
	}
	return service.NewAuthService(userRepo, sessionRepo, cfg)
}

func TestAuthService_Register(t *testing.T) {
	userRepo := newMockUserRepo()
	sessionRepo := newMockAuthSessionRepo()
	authSvc := newTestAuthServiceWithRepos(userRepo, sessionRepo)

	user, tokens, refreshToken, err := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "test@test.com", DisplayName: "TestUser", Password: "password123",
	}, service.AuthSessionMetadata{UserAgent: "test-suite", IPAddress: "127.0.0.1"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user == nil || user.Email != "test@test.com" {
		t.Error("expected user with email test@test.com")
	}
	if tokens == nil || tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if len(sessionRepo.sessions) != 1 {
		t.Errorf("expected 1 stored auth session, got %d", len(sessionRepo.sessions))
	}

	_, _, _, err = authSvc.Register(context.Background(), domain.UserCreate{
		Email: "test@test.com", DisplayName: "TestUser2", Password: "pass45678",
	}, service.AuthSessionMetadata{})
	if err != service.ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	authSvc := newTestAuthService()

	_, _, _, err := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "user@test.com", DisplayName: "User", Password: "mypassword",
	}, service.AuthSessionMetadata{})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, tokens, refreshToken, err := authSvc.Login(context.Background(), domain.UserLogin{
		Email: "user@test.com", Password: "mypassword",
	}, service.AuthSessionMetadata{})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tokens == nil || tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	_, _, _, err = authSvc.Login(context.Background(), domain.UserLogin{
		Email: "user@test.com", Password: "wrongpassword",
	}, service.AuthSessionMetadata{})
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	_, _, _, err = authSvc.Login(context.Background(), domain.UserLogin{
		Email: "noexist@test.com", Password: "pass",
	}, service.AuthSessionMetadata{})
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for non-existent user, got %v", err)
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	authSvc := newTestAuthService()

	_, tokens, _, _ := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "tok@test.com", DisplayName: "TokUser", Password: "pass12345",
	}, service.AuthSessionMetadata{})

	claims, err := authSvc.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID == "" {
		t.Error("expected non-empty userID in claims")
	}
	if claims.SessionID == "" {
		t.Error("expected non-empty sessionID in claims")
	}

	_, err = authSvc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestAuthService_RefreshRotatesSession(t *testing.T) {
	userRepo := newMockUserRepo()
	sessionRepo := newMockAuthSessionRepo()
	authSvc := newTestAuthServiceWithRepos(userRepo, sessionRepo)

	_, _, refreshToken, err := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "refresh@test.com", DisplayName: "Refresh", Password: "pass12345",
	}, service.AuthSessionMetadata{})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	tokens, nextRefreshToken, err := authSvc.Refresh(context.Background(), refreshToken, service.AuthSessionMetadata{})
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	if tokens == nil || tokens.AccessToken == "" {
		t.Fatal("expected rotated access token")
	}
	if nextRefreshToken == "" || nextRefreshToken == refreshToken {
		t.Fatal("expected rotated refresh token")
	}
	if len(sessionRepo.sessions) != 2 {
		t.Fatalf("expected original and rotated sessions to be stored, got %d", len(sessionRepo.sessions))
	}

	_, _, err = authSvc.Refresh(context.Background(), refreshToken, service.AuthSessionMetadata{})
	if err != service.ErrUnauthorized {
		t.Fatalf("expected old refresh token reuse to be rejected, got %v", err)
	}
}
