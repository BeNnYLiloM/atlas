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

	cfg := config.Load()
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	if err := database.RunMigrations(cfg.Database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := postgres.NewUserRepo(db)
	authSessionRepo := postgres.NewAuthSessionRepo(db)
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
	projectRepo := postgres.NewProjectRepo(db)
	dmChannelRepo := postgres.NewDMChannelRepo(db)

	wsHub := ws.NewHub(userRepo)
	go wsHub.Run()

	authService := service.NewAuthService(userRepo, authSessionRepo, cfg.JWT)
	workspaceService := service.NewWorkspaceService(workspaceRepo, channelRepo, roleRepo, projectRepo)
	categoryService := service.NewChannelCategoryService(channelCategoryRepo, categoryPermRepo, channelRepo, workspaceRepo, roleRepo, projectRepo)
	roleService := service.NewWorkspaceRoleService(roleRepo, workspaceRepo)
	channelService := service.NewChannelService(channelRepo, workspaceRepo, channelMemberRepo, channelPermRepo, roleRepo, projectRepo, dmChannelRepo)
	messageService := service.NewMessageService(messageRepo, channelRepo, workspaceRepo, channelMemberRepo, channelPermRepo, roleRepo, projectRepo, dmChannelRepo)
	projectService := service.NewProjectService(projectRepo, workspaceRepo, roleRepo, channelRepo, channelPermRepo, channelMemberRepo)
	liveKitService := service.NewLiveKitService(cfg.LiveKit)
	searchService := service.NewSearchService(searchRepo, workspaceRepo, channelRepo, roleRepo, channelPermRepo, projectRepo, dmChannelRepo)
	reactionService := service.NewReactionService(reactionRepo, wsHub)
	taskService := service.NewTaskService(taskRepo, workspaceRepo, messageRepo, channelRepo, roleRepo, channelPermRepo, projectRepo, dmChannelRepo)
	dmService := service.NewDMService(dmChannelRepo, userRepo, workspaceRepo)

	minioStorage, minioErr := storage.NewMinIOStorage(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.UseSSL,
		cfg.MinIO.PublicURL,
	)
	var fileService *service.FileService
	if minioErr != nil {
		log.Printf("Warning: MinIO unavailable: %v. File upload disabled.", minioErr)
	} else {
		fileService = service.NewFileService(fileRepo, minioStorage)
	}

	router := gin.Default()
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	authMiddleware := middleware.AuthMiddleware(authService)

	authHandler := handler.NewAuthHandler(authService, fileService, cfg.JWT, wsHub)
	authHandler.RegisterRoutes(api, authMiddleware)

	channelHandler := handler.NewChannelHandler(channelService, projectService, wsHub)
	messageHandler := handler.NewMessageHandler(messageService, channelService, projectService, wsHub)
	messageHandler.RegisterRoutes(api, authMiddleware)
	channelHandler.RegisterWithMessages(api, authMiddleware, messageHandler)

	workspaceHandler := handler.NewWorkspaceHandler(workspaceService, fileService, wsHub)
	roleHandler := handler.NewWorkspaceRoleHandler(roleService, channelService, projectService, wsHub)
	categoryHandler := handler.NewChannelCategoryHandler(categoryService, channelPermRepo, wsHub)
	workspaceHandler.RegisterRoutes(api, authMiddleware, channelHandler, roleHandler, categoryHandler)

	searchHandler := handler.NewSearchHandler(searchService)
	reactionHandler := handler.NewReactionHandler(reactionService)
	taskHandler := handler.NewTaskHandler(taskService)
	callsHandler := handler.NewCallsHandler(liveKitService, authService, channelService, wsHub)
	projectHandler := handler.NewProjectHandler(projectService, channelService, channelCategoryRepo, fileService, wsHub)
	projectHandler.RegisterRoutes(api, authMiddleware)

	dmHandler := handler.NewDMHandler(dmService)
	dmRateLimiter := middleware.NewRateLimiter(20, time.Hour)

	protected := api.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/dm", dmHandler.List)
		protected.POST("/dm", dmRateLimiter, dmHandler.Open)
		protected.GET("/search", searchHandler.Search)
		protected.POST("/messages/:id/reactions", reactionHandler.Add)
		protected.DELETE("/messages/:id/reactions/:emoji", reactionHandler.Remove)
		protected.GET("/messages/:id/reactions", reactionHandler.GetReactions)
		protected.POST("/tasks", taskHandler.Create)
		protected.GET("/tasks", taskHandler.List)
		protected.PATCH("/tasks/:id", taskHandler.Update)
		protected.DELETE("/tasks/:id", taskHandler.Delete)
		protected.POST("/calls/join", callsHandler.JoinCall)
		protected.POST("/calls/signal", callsHandler.SignalCall)

		if fileService != nil {
			fileHandler := handler.NewFileHandler(fileService)
			protected.POST("/files/upload", fileHandler.Upload)
			protected.GET("/files/:id", fileHandler.GetByID)
			protected.DELETE("/files/:id", fileHandler.Delete)
		}
	}

	wsHandler := ws.NewHandler(wsHub, authService, channelService, cfg.Server.AllowedOrigins)
	wsHandler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

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
