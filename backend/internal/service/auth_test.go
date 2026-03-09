package service_test

import (
	"context"
	"testing"

	"github.com/your-org/atlas/backend/internal/config"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/service"
)

// mockUserRepo - мок-репозиторий пользователей
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
		return nil, nil // репозиторий возвращает nil, nil если не найден
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

func newTestAuthService() *service.AuthService {
	repo := newMockUserRepo()
	cfg := config.JWTConfig{Secret: "test-secret-key", ExpireHour: 24}
	return service.NewAuthService(repo, cfg)
}

func newTestAuthServiceWithRepo(repo *mockUserRepo) *service.AuthService {
	cfg := config.JWTConfig{Secret: "test-secret-key", ExpireHour: 24}
	return service.NewAuthService(repo, cfg)
}

func TestAuthService_Register(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := newTestAuthServiceWithRepo(repo)

	user, tokens, err := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "test@test.com", DisplayName: "TestUser", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user == nil || user.Email != "test@test.com" {
		t.Error("expected user with email test@test.com")
	}
	if tokens == nil || tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	// Повторная регистрация должна вернуть ошибку
	_, _, err = authSvc.Register(context.Background(), domain.UserCreate{
		Email: "test@test.com", DisplayName: "TestUser2", Password: "pass456",
	})
	if err != service.ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := newTestAuthServiceWithRepo(repo)

	// Сначала регистрируем
	_, _, err := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "user@test.com", DisplayName: "User", Password: "mypassword",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Успешный логин
	_, tokens, err := authSvc.Login(context.Background(), domain.UserLogin{
		Email: "user@test.com", Password: "mypassword",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tokens == nil || tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	// Неверный пароль
	_, _, err = authSvc.Login(context.Background(), domain.UserLogin{
		Email: "user@test.com", Password: "wrongpassword",
	})
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// Несуществующий пользователь
	_, _, err = authSvc.Login(context.Background(), domain.UserLogin{
		Email: "noexist@test.com", Password: "pass",
	})
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for non-existent user, got %v", err)
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := newTestAuthServiceWithRepo(repo)

	_, tokens, _ := authSvc.Register(context.Background(), domain.UserCreate{
		Email: "tok@test.com", DisplayName: "TokUser", Password: "pass123",
	})

	// Валидный токен
	claims, err := authSvc.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID == "" {
		t.Error("expected non-empty userID in claims")
	}

	// Невалидный токен
	_, err = authSvc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}
