package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/your-org/atlas/backend/internal/config"
    "github.com/your-org/atlas/backend/internal/domain"
    "github.com/your-org/atlas/backend/internal/repository/postgres"
    "github.com/your-org/atlas/backend/internal/service"
    "github.com/your-org/atlas/backend/pkg/database"
)

func main() {
    cfg := config.Load()
    ctx := context.Background()

    db, err := database.NewPostgresPool(ctx, cfg.Database)
    if err != nil {
        log.Fatalf("db: %v", err)
    }
    defer db.Close()

    userRepo := postgres.NewUserRepo(db)
    sessionRepo := postgres.NewAuthSessionRepo(db)
    authSvc := service.NewAuthService(userRepo, sessionRepo, cfg.JWT)

    email := fmt.Sprintf("refresh-service-%d@test.local", time.Now().UnixNano())
    user, tokens, refreshToken, err := authSvc.Register(ctx, domain.UserCreate{
        Email: email,
        DisplayName: "Refresh Service",
        Password: "password123",
    }, service.AuthSessionMetadata{UserAgent: "debug", IPAddress: "127.0.0.1"})
    fmt.Printf("register user=%s access=%t refresh=%t err=%v\n", user.Email, tokens != nil && tokens.AccessToken != "", refreshToken != "", err)
    if err != nil {
        return
    }

    nextTokens, nextRefreshToken, err := authSvc.Refresh(ctx, refreshToken, service.AuthSessionMetadata{UserAgent: "debug", IPAddress: "127.0.0.1"})
    fmt.Printf("refresh access=%t nextRefresh=%t err=%T %v\n", nextTokens != nil && nextTokens.AccessToken != "", nextRefreshToken != "", err, err)
}
