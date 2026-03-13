package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/your-org/atlas/backend/internal/config"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
)

const refreshTokenEntropyBytes = 32

type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.AuthSessionRepository
	jwtConfig   config.JWTConfig
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

type AuthSessionMetadata struct {
	UserAgent string
	IPAddress string
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.AuthSessionRepository, jwtConfig config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtConfig:   jwtConfig,
	}
}

// Register создает нового пользователя.
func (s *AuthService) Register(ctx context.Context, input domain.UserCreate, meta AuthSessionMetadata) (*domain.User, *TokenPair, string, error) {
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, "", err
	}
	if existing != nil {
		return nil, nil, "", ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, "", err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  input.DisplayName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, "", err
	}

	tokens, refreshToken, err := s.createSession(ctx, user, "", meta)
	if err != nil {
		return nil, nil, "", err
	}

	return user, tokens, refreshToken, nil
}

// Login аутентифицирует пользователя.
func (s *AuthService) Login(ctx context.Context, input domain.UserLogin, meta AuthSessionMetadata) (*domain.User, *TokenPair, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, "", err
	}
	if user == nil {
		return nil, nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, "", ErrInvalidCredentials
	}

	tokens, refreshToken, err := s.createSession(ctx, user, "", meta)
	if err != nil {
		return nil, nil, "", err
	}

	return user, tokens, refreshToken, nil
}

// Refresh обновляет access token и ротирует refresh session.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta AuthSessionMetadata) (*TokenPair, string, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, "", ErrUnauthorized
	}

	hash := hashToken(refreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, "", err
	}
	if session == nil {
		return nil, "", ErrUnauthorized
	}

	now := time.Now().UTC()
	if session.RevokedAt != nil {
		if session.ReplacedBySessionID != nil {
			log.Printf("[AUTH] refresh token reuse detected: session=%s user=%s", session.ID, session.UserID)
			if revokeErr := s.sessionRepo.RevokeFamily(ctx, session.FamilyID); revokeErr != nil {
				log.Printf("[AUTH] failed to revoke compromised session family %s: %v", session.FamilyID, revokeErr)
			}
		}
		return nil, "", ErrUnauthorized
	}
	if !session.ExpiresAt.After(now) {
		if revokeErr := s.sessionRepo.RevokeByID(ctx, session.ID); revokeErr != nil {
			log.Printf("[AUTH] failed to revoke expired session %s: %v", session.ID, revokeErr)
		}
		return nil, "", ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrUnauthorized
	}

	tokens, nextRefreshToken, nextSession, err := s.generateSessionArtifacts(user, session.FamilyID, meta)
	if err != nil {
		return nil, "", err
	}

	if err := s.sessionRepo.Rotate(ctx, session.ID, nextSession); err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", ErrUnauthorized
		}
		return nil, "", err
	}

	return tokens, nextRefreshToken, nil
}

// Logout завершает текущую refresh-сессию.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}

	hash := hashToken(refreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}

	return s.sessionRepo.RevokeByID(ctx, session.ID)
}

// LogoutAll завершает все активные сессии пользователя.
func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrUnauthorized
	}
	return s.sessionRepo.RevokeAllByUserID(ctx, userID)
}

// ValidateToken проверяет JWT токен и возвращает claims.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return []byte(s.jwtConfig.Secret), nil
	}, jwt.WithAudience(s.jwtConfig.Audience), jwt.WithIssuer(s.jwtConfig.Issuer))
	if err != nil {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

// GetUserByID возвращает пользователя по ID.
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile обновляет профиль текущего пользователя.
func (s *AuthService) UpdateProfile(ctx context.Context, userID string, input domain.UserUpdate) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if input.DisplayName != nil {
		displayName := strings.TrimSpace(*input.DisplayName)
		if len(displayName) < 2 || len(displayName) > 100 {
			return nil, ErrInvalidProfile
		}
		user.DisplayName = displayName
	}

	if input.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*input.AvatarURL)
		if avatarURL == "" {
			user.AvatarURL = nil
		} else {
			user.AvatarURL = &avatarURL
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword меняет пароль пользователя после проверки текущего.
func (s *AuthService) ChangePassword(ctx context.Context, userID string, input domain.UserChangePassword) error {
	if strings.TrimSpace(input.CurrentPassword) == "" || strings.TrimSpace(input.NewPassword) == "" {
		return ErrInvalidProfile
	}
	if len(input.NewPassword) < 8 {
		return ErrWeakPassword
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return err
	}

	// Отзываем все сессии кроме текущей не делаем — пользователь сам решает.
	// Но логируем факт смены пароля.
	log.Printf("[AUTH] password changed for user=%s", userID)
	return nil
}

// UpdateStatus обновляет статус и кастомный статус пользователя.
func (s *AuthService) UpdateStatus(ctx context.Context, userID string, status domain.UserStatus, customStatus *string) (*domain.User, error) {
	validStatuses := map[domain.UserStatus]bool{
		domain.UserStatusOnline:  true,
		domain.UserStatusAway:    true,
		domain.UserStatusDND:     true,
		domain.UserStatusOffline: true,
	}
	if !validStatuses[status] {
		return nil, ErrInvalidProfile
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.Status = status
	if customStatus != nil {
		trimmed := strings.TrimSpace(*customStatus)
		if len(trimmed) > 100 {
			return nil, ErrInvalidProfile
		}
		if trimmed == "" {
			user.CustomStatus = nil
		} else {
			user.CustomStatus = &trimmed
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteAccount удаляет аккаунт пользователя после проверки пароля.
func (s *AuthService) DeleteAccount(ctx context.Context, userID, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	// Сначала отзываем все сессии
	if err := s.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		log.Printf("[AUTH] failed to revoke sessions before account deletion user=%s: %v", userID, err)
	}

	return s.userRepo.DeleteByID(ctx, userID)
}

// GetUserByEmail возвращает пользователя по email.
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) RefreshCookieName() string {
	return s.jwtConfig.RefreshCookieName
}

func (s *AuthService) RefreshCookieDomain() string {
	return s.jwtConfig.RefreshCookieDomain
}

func (s *AuthService) RefreshCookieSecure() bool {
	return s.jwtConfig.RefreshCookieSecure
}

func (s *AuthService) RefreshCookieMaxAgeSeconds() int {
	return int((time.Hour * 24 * time.Duration(s.jwtConfig.RefreshTokenTTLDays)).Seconds())
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User, familyID string, meta AuthSessionMetadata) (*TokenPair, string, error) {
	tokens, refreshToken, session, err := s.generateSessionArtifacts(user, familyID, meta)
	if err != nil {
		return nil, "", err
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", err
	}

	return tokens, refreshToken, nil
}

func (s *AuthService) generateSessionArtifacts(user *domain.User, familyID string, meta AuthSessionMetadata) (*TokenPair, string, *domain.AuthSession, error) {
	now := time.Now().UTC()
	if familyID == "" {
		familyID = uuid.NewString()
	}

	sessionID := uuid.NewString()
	accessExpiry := now.Add(time.Minute * time.Duration(s.jwtConfig.AccessTokenTTLMinutes))
	refreshExpiry := now.Add(time.Hour * 24 * time.Duration(s.jwtConfig.RefreshTokenTTLDays))

	accessToken, err := s.generateAccessToken(user, sessionID, accessExpiry, now)
	if err != nil {
		return nil, "", nil, err
	}

	refreshToken, refreshHash, err := generateOpaqueToken()
	if err != nil {
		return nil, "", nil, err
	}

	session := &domain.AuthSession{
		ID:               sessionID,
		FamilyID:         familyID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        clampString(meta.UserAgent, 512),
		IPAddress:        clampString(meta.IPAddress, 128),
		CreatedAt:        now,
		ExpiresAt:        refreshExpiry,
		LastUsedAt:       now,
	}

	return &TokenPair{
		AccessToken: accessToken,
		ExpiresAt:   accessExpiry.Unix(),
	}, refreshToken, session, nil
}

func (s *AuthService) ListActiveSessions(ctx context.Context, userID string) ([]*domain.AuthSession, error) {
	return s.sessionRepo.ListActiveByUserID(ctx, userID)
}

// RevokeSession отзывает конкретную сессию пользователя.
// Возвращает false если сессия не найдена или принадлежит другому пользователю.
func (s *AuthService) RevokeSession(ctx context.Context, sessionID, userID string) (bool, error) {
	return s.sessionRepo.RevokeByIDForUser(ctx, sessionID, userID)
}

func (s *AuthService) generateAccessToken(user *domain.User, sessionID string, expiresAt, issuedAt time.Time) (string, error) {
	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.jwtConfig.Issuer,
			Audience:  jwt.ClaimStrings{s.jwtConfig.Audience},
			Subject:   user.ID,
			ID:        sessionID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

func generateOpaqueToken() (string, string, error) {
	buf := make([]byte, refreshTokenEntropyBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(buf)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func clampString(value string, maxLen int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return trimmed[:maxLen]
}
