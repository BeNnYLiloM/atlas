package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/config"
	"github.com/your-org/atlas/backend/internal/repository/postgres"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/handler"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/ws"
	"github.com/your-org/atlas/backend/pkg/database"
	"github.com/your-org/atlas/backend/pkg/storage"
)

func main() {
	log.Println("Starting Atlas server...")

	// Загружаем конфигурацию
	cfg := config.Load()

	// Устанавливаем режим Gin
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Подключаемся к PostgreSQL
	ctx := context.Background()
	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// Применяем миграции
	if err := database.RunMigrations(cfg.Database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализируем репозитории
	userRepo := postgres.NewUserRepo(db)
	workspaceRepo := postgres.NewWorkspaceRepo(db)
	channelRepo := postgres.NewChannelRepo(db)
	channelCategoryRepo := postgres.NewChannelCategoryRepo(db)
	categoryPermRepo := postgres.NewCategoryPermissionRepo(db)
	channelMemberRepo := postgres.NewChannelMemberRepo(db)
	channelPermRepo := postgres.NewChannelPermissionRepo(db)
	roleRepo := postgres.NewWorkspaceRoleRepo(db)
	messageRepo := postgres.NewMessageRepo(db)
	fileRepo := postgres.NewFileRepository(db)
	searchRepo := postgres.NewSearchRepository(db)
	reactionRepo := postgres.NewReactionRepository(db)
	taskRepo := postgres.NewTaskRepository(db)

	// Инициализируем WebSocket hub (до сервисов, т.к. некоторые зависят от него)
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Инициализируем сервисы
	authService := service.NewAuthService(userRepo, cfg.JWT)
	workspaceService := service.NewWorkspaceService(workspaceRepo, channelRepo, roleRepo)
	categoryService := service.NewChannelCategoryService(channelCategoryRepo, categoryPermRepo, channelRepo, workspaceRepo, roleRepo)
	roleService := service.NewWorkspaceRoleService(roleRepo, workspaceRepo)
	channelService := service.NewChannelService(channelRepo, workspaceRepo, channelMemberRepo, channelPermRepo, roleRepo)
	messageService := service.NewMessageService(messageRepo, channelRepo, workspaceRepo, channelMemberRepo)
	liveKitService := service.NewLiveKitService(cfg.LiveKit)
	searchService := service.NewSearchService(searchRepo)
	reactionService := service.NewReactionService(reactionRepo, wsHub)
	taskService := service.NewTaskService(taskRepo)

	// MinIO storage (не блокируем старт если MinIO недоступен)
	minioStorage, minioErr := storage.NewMinIOStorage(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.UseSSL,
	)
	var fileService *service.FileService
	if minioErr != nil {
		log.Printf("Warning: MinIO unavailable: %v. File upload disabled.", minioErr)
	} else {
		fileService = service.NewFileService(fileRepo, minioStorage)
	}

	// Создаем роутер
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API группа
	api := router.Group("/api/v1")

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(authService)

	// Регистрируем HTTP handlers
	authHandler := handler.NewAuthHandler(authService)
	authHandler.RegisterRoutes(api, authMiddleware)

	channelHandler := handler.NewChannelHandler(channelService, wsHub)
	messageHandler := handler.NewMessageHandler(messageService, channelService, wsHub)
	messageHandler.RegisterRoutes(api, authMiddleware)

	channelHandler.RegisterWithMessages(api, authMiddleware, messageHandler)

	workspaceHandler := handler.NewWorkspaceHandler(workspaceService, fileService, wsHub)
	roleHandler := handler.NewWorkspaceRoleHandler(roleService, channelService, wsHub)
	categoryHandler := handler.NewChannelCategoryHandler(categoryService, channelPermRepo, wsHub)
	workspaceHandler.RegisterRoutes(api, authMiddleware, channelHandler, roleHandler, categoryHandler)

	// Новые handlers
	searchHandler := handler.NewSearchHandler(searchService)
	reactionHandler := handler.NewReactionHandler(reactionService)
	taskHandler := handler.NewTaskHandler(taskService)
	callsHandler := handler.NewCallsHandler(liveKitService, authService)

	protected := api.Group("")
	protected.Use(authMiddleware)
	{
		// Поиск
		protected.GET("/search", searchHandler.Search)

		// Реакции
		protected.POST("/messages/:id/reactions", reactionHandler.Add)
		protected.DELETE("/messages/:id/reactions/:emoji", reactionHandler.Remove)
		protected.GET("/messages/:id/reactions", reactionHandler.GetReactions)

		// Задачи
		protected.POST("/tasks", taskHandler.Create)
		protected.GET("/tasks", taskHandler.List)
		protected.PATCH("/tasks/:id", taskHandler.Update)
		protected.DELETE("/tasks/:id", taskHandler.Delete)

		// Звонки (LiveKit)
		protected.POST("/calls/join", callsHandler.JoinCall)

		// Файлы
		if fileService != nil {
			fileHandler := handler.NewFileHandler(fileService)
			protected.POST("/files/upload", fileHandler.Upload)
			protected.GET("/files/:id", fileHandler.GetByID)
			protected.DELETE("/files/:id", fileHandler.Delete)
		}
	}

	// WebSocket handler
	wsHandler := ws.NewHandler(wsHub, authService)
	wsHandler.RegisterRoutes(router)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
