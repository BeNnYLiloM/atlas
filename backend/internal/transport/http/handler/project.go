package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/internal/repository"
	"github.com/your-org/atlas/backend/internal/service"
	"github.com/your-org/atlas/backend/internal/transport/http/middleware"
	"github.com/your-org/atlas/backend/internal/transport/http/response"
	"github.com/your-org/atlas/backend/internal/transport/ws"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	channelService *service.ChannelService
	categoryRepo   repository.ChannelCategoryRepository
	fileService    *service.FileService
	wsHub          *ws.Hub
}

func NewProjectHandler(
	projectService *service.ProjectService,
	channelService *service.ChannelService,
	categoryRepo repository.ChannelCategoryRepository,
	fileService *service.FileService,
	wsHub *ws.Hub,
) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		channelService: channelService,
		categoryRepo:   categoryRepo,
		fileService:    fileService,
		wsHub:          wsHub,
	}
}

func (h *ProjectHandler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc) {
	projects := rg.Group("/workspaces/:id/projects")
	projects.Use(auth)
	{
		projects.GET("", h.List)
		projects.POST("", h.Create)
	}

	project := rg.Group("/projects/:projectId")
	project.Use(auth)
	{
		project.GET("", h.GetByID)
		project.PATCH("", h.Update)
		project.DELETE("", h.Delete)
		project.POST("/archive", h.Archive)
		project.DELETE("/archive", h.Unarchive)
		project.POST("/icon", h.UploadIcon)
		project.GET("/channels", h.ListChannels)
		project.GET("/categories", h.ListCategories)
		project.GET("/members", h.ListMembers)
		project.POST("/members", h.AddMember)
		project.DELETE("/members/:userId", h.RemoveMember)
		project.POST("/members/:userId/lead", h.SetLead)
		project.DELETE("/members/:userId/lead", h.UnsetLead)
	}
}

func (h *ProjectHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	projects, err := h.projectService.List(c.Request.Context(), workspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, projects)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var input domain.ProjectCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project, err := h.projectService.Create(c.Request.Context(), workspaceID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Уведомляем только участников с view_all_projects + создателя через BroadcastToUsers
	recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, workspaceID)
	h.wsHub.BroadcastToUsers(recipientIDs, "project_created", project)

	response.Created(c, project)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	project, err := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, project)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	var input domain.ProjectUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project, err := h.projectService.Update(c.Request.Context(), projectID, input, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// workspaceID берём из объекта проекта из БД (не из запроса)
	recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)
	h.wsHub.BroadcastToUsers(recipientIDs, "project_updated", project)

	response.Success(c, project)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	if c.Query("force") != "true" {
		response.BadRequest(c, "deletion requires ?force=true")
		return
	}

	// Получаем проект до удаления чтобы иметь workspaceID и список получателей
	project, err := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)

	if err := h.projectService.Delete(c.Request.Context(), projectID, userID, true); err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToUsers(recipientIDs, "project_deleted", gin.H{"project_id": projectID})
	response.NoContent(c)
}

func (h *ProjectHandler) Archive(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	if err := h.projectService.Archive(c.Request.Context(), projectID, userID); err != nil {
		response.Error(c, err)
		return
	}

	project, _ := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if project != nil {
		recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)
		h.wsHub.BroadcastToUsers(recipientIDs, "project_updated", project)
	}

	response.NoContent(c)
}

func (h *ProjectHandler) Unarchive(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	if err := h.projectService.Unarchive(c.Request.Context(), projectID, userID); err != nil {
		response.Error(c, err)
		return
	}

	project, _ := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if project != nil {
		recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)
		h.wsHub.BroadcastToUsers(recipientIDs, "project_updated", project)
	}

	response.NoContent(c)
}

func (h *ProjectHandler) ListMembers(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	members, err := h.projectService.GetMembers(c.Request.Context(), projectID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, members)
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	projectID := c.Param("projectId")
	actorID := middleware.GetUserID(c)

	var input domain.ProjectMemberAdd
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.projectService.AddMember(c.Request.Context(), projectID, input.UserID, actorID); err != nil {
		response.Error(c, err)
		return
	}

	project, _ := h.projectService.GetByID(c.Request.Context(), projectID, actorID)
	if project != nil {
		recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)
		h.wsHub.BroadcastToUsers(recipientIDs, "project_member_added", gin.H{
			"project_id": projectID,
			"user_id":    input.UserID,
		})
	}

	response.NoContent(c)
}

func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	projectID := c.Param("projectId")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	project, err := h.projectService.GetByID(c.Request.Context(), projectID, actorID)
	if err != nil {
		response.Error(c, err)
		return
	}
	recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)

	if err := h.projectService.RemoveMember(c.Request.Context(), projectID, targetUserID, actorID); err != nil {
		response.Error(c, err)
		return
	}

	h.wsHub.BroadcastToUsers(recipientIDs, "project_member_removed", gin.H{
		"project_id": projectID,
		"user_id":    targetUserID,
	})

	response.NoContent(c)
}

func (h *ProjectHandler) SetLead(c *gin.Context) {
	projectID := c.Param("projectId")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	if err := h.projectService.SetLead(c.Request.Context(), projectID, targetUserID, actorID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *ProjectHandler) UnsetLead(c *gin.Context) {
	projectID := c.Param("projectId")
	targetUserID := c.Param("userId")
	actorID := middleware.GetUserID(c)

	if err := h.projectService.UnsetLead(c.Request.Context(), projectID, targetUserID, actorID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *ProjectHandler) ListCategories(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	project, err := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	all, err := h.categoryRepo.GetByWorkspaceID(c.Request.Context(), project.WorkspaceID)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]*domain.ChannelCategory, 0)
	for _, cat := range all {
		if cat.ProjectID != nil && *cat.ProjectID == projectID {
			result = append(result, cat)
		}
	}
	response.Success(c, result)
}

func (h *ProjectHandler) UploadIcon(c *gin.Context) {
	if h.fileService == nil {
		response.BadRequest(c, "file storage unavailable")
		return
	}

	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	// HIGH-1: проверяем права ДО загрузки файла
	if _, err := h.projectService.GetByID(c.Request.Context(), projectID, userID); err != nil {
		response.Error(c, err)
		return
	}
	// canManageProject проверяется внутри Update, но GetByID уже подтверждает membership.
	// Дополнительно вызываем canManage через Update ниже.

	file, header, err := c.Request.FormFile("icon")
	if err != nil {
		response.BadRequest(c, "icon file is required")
		return
	}
	defer file.Close()

	// HIGH-2: валидируем Content-Type — принимаем только изображения
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		response.BadRequest(c, "unsupported file type: only jpeg, png, gif, webp are allowed")
		return
	}

	uploaded, err := h.fileService.Upload(c.Request.Context(), userID, header.Filename, file, header.Size, contentType)
	if err != nil {
		response.BadRequest(c, fmt.Sprintf("upload failed: %v", err))
		return
	}

	project, err := h.projectService.Update(c.Request.Context(), projectID, domain.ProjectUpdate{
		IconURL: &uploaded.URL,
	}, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	recipientIDs, _ := h.projectService.GetProjectMembersAndViewAll(c.Request.Context(), project.ID, project.WorkspaceID)
	h.wsHub.BroadcastToUsers(recipientIDs, "project_updated", project)
	response.Success(c, project)
}

func (h *ProjectHandler) ListChannels(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := middleware.GetUserID(c)

	// Получаем проект чтобы взять workspaceID
	project, err := h.projectService.GetByID(c.Request.Context(), projectID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	channels, err := h.channelService.GetByProjectIDWithUnread(c.Request.Context(), projectID, project.WorkspaceID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if channels == nil {
		channels = []*domain.ChannelWithUnread{}
	}
	response.Success(c, channels)
}
