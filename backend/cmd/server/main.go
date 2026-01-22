package main

import (
	"os"
	"time"

	"github.com/ekf/one-on-one-backend/internal/config"
	"github.com/ekf/one-on-one-backend/internal/handlers"
	"github.com/ekf/one-on-one-backend/internal/middleware"
	"github.com/ekf/one-on-one-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

func main() {
	// Load config
	cfg := config.Load()

	// Create handler
	h := handlers.NewHandler(cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit:    100 * 1024 * 1024, // 100MB for audio files
		ReadTimeout:  120 * time.Second, // Увеличен таймаут для синхронизации календаря
		WriteTimeout: 120 * time.Second,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// CORS configuration
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173,http://localhost:3000,https://one-on-one-front.up.railway.app"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-CSRF-Token",
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	// Global API rate limiting (100 req/min per IP)
	app.Use(middleware.APIRateLimiter())

	// Health endpoints
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "EKF Hub API",
			"version": "1.0.0",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		status := "healthy"
		statusCode := 200

		// Check database connection by attempting a simple query
		dbStatus := "connected"
		if h.DB == nil {
			dbStatus = "not configured"
			status = "degraded"
		} else {
			// Try to execute a simple query to verify DB connectivity
			var employees []struct{ ID string }
			err := h.DB.From("employees").Select("id").Limit(1).Execute(&employees)
			if err != nil {
				dbStatus = "disconnected"
				status = "unhealthy"
				statusCode = 503
			}
		}

		// Check EWS configuration
		ewsStatus := "not configured"
		if cfg.EWSURL != "" {
			ewsStatus = "configured"
		}

		// Check AI services
		aiStatus := "not configured"
		if cfg.OpenAIKey != "" || cfg.AnthropicKey != "" {
			aiStatus = "configured"
		}

		response := fiber.Map{
			"status":     status,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"database":   dbStatus,
			"ews":        ewsStatus,
			"ews_url":    cfg.EWSURL,
			"ai_service": aiStatus,
		}

		return c.Status(statusCode).JSON(response)
	})

	// API routes
	api := app.Group("/api/v1")

	// Public routes (no JWT required)
	publicAPI := api.Group("")
	// Auth endpoint with stricter rate limiting (5 req/min per IP to prevent brute force)
	publicAPI.Post("/ad/authenticate", middleware.AuthRateLimiter(), h.AuthenticateAD)
	publicAPI.Get("/connector/status", h.ConnectorStatus)

	// Protected routes (JWT required)
	protectedAPI := api.Group("", middleware.JWTAuth(h.JWT))

	// CSRF token endpoint (must be called before state-changing requests)
	protectedAPI.Get("/csrf-token", middleware.CSRFTokenHandler)

	// Apply CSRF protection to all state-changing routes
	protectedAPI.Use(middleware.CSRFProtection())

	// Employees
	protectedAPI.Get("/employees", h.ListEmployees)
	protectedAPI.Post("/employees", h.CreateEmployee)
	protectedAPI.Get("/employees/:id", h.GetEmployee)
	protectedAPI.Put("/employees/:id", h.UpdateEmployee)
	protectedAPI.Delete("/employees/:id", h.DeleteEmployee)
	protectedAPI.Get("/employees/:id/dossier", h.GetEmployeeDossier)
	protectedAPI.Get("/employees/my-team", h.GetMyTeam)

	// Projects
	protectedAPI.Get("/projects", h.ListProjects)
	protectedAPI.Post("/projects", h.CreateProject)
	protectedAPI.Get("/projects/:id", h.GetProject)
	protectedAPI.Put("/projects/:id", h.UpdateProject)
	protectedAPI.Delete("/projects/:id", h.DeleteProject)

	// Meetings
	protectedAPI.Get("/meetings", h.ListMeetings)
	protectedAPI.Post("/meetings", h.CreateMeeting)
	protectedAPI.Get("/meetings/:id", h.GetMeeting)
	protectedAPI.Get("/meeting-categories", h.ListMeetingCategories)
	protectedAPI.Get("/ai/status", h.AIStatus)
	protectedAPI.Post("/process-meeting", h.ProcessMeeting)

	// Tasks
	protectedAPI.Get("/tasks", h.ListTasks)
	protectedAPI.Post("/tasks", h.CreateTask)
	protectedAPI.Get("/tasks/:id", h.GetTask)
	protectedAPI.Put("/tasks/:id", h.UpdateTask)
	protectedAPI.Delete("/tasks/:id", h.DeleteTask)
	// Task dependencies
	protectedAPI.Get("/tasks/:id/dependencies", h.GetTaskDependencies)
	protectedAPI.Post("/tasks/:id/dependencies", h.AddTaskDependency)
	protectedAPI.Delete("/tasks/:id/dependencies/:dep_id", h.RemoveTaskDependency)
	protectedAPI.Get("/tasks/:id/blocked", h.IsTaskBlocked)
	// Time entries
	protectedAPI.Get("/tasks/:id/time-entries", h.ListTimeEntries)
	protectedAPI.Post("/tasks/:id/time-entries", h.CreateTimeEntry)
	protectedAPI.Put("/tasks/:id/time-entries/:entry_id", h.UpdateTimeEntry)
	protectedAPI.Delete("/tasks/:id/time-entries/:entry_id", h.DeleteTimeEntry)
	protectedAPI.Get("/tasks/:id/resources", h.GetTaskResourceSummary)
	// My time entries
	protectedAPI.Get("/time-entries/me", h.GetMyTimeEntries)

	// Kanban
	protectedAPI.Get("/kanban", h.GetKanban)
	protectedAPI.Put("/kanban/move", h.MoveTaskKanban)

	// Workflows
	protectedAPI.Get("/workflows/me", h.GetWorkflowForUser)
	protectedAPI.Get("/workflows", h.ListWorkflowModes)
	protectedAPI.Get("/workflows/departments", h.ListDepartmentWorkflows)
	protectedAPI.Post("/workflows/departments", h.SetDepartmentWorkflow)

	// Versions/Releases (JIRA-like)
	protectedAPI.Get("/versions", h.ListVersions)
	protectedAPI.Post("/versions", h.CreateVersion)
	protectedAPI.Get("/versions/:id", h.GetVersion)
	protectedAPI.Put("/versions/:id", h.UpdateVersion)
	protectedAPI.Delete("/versions/:id", h.DeleteVersion)
	protectedAPI.Post("/versions/:id/release", h.ReleaseVersion)
	protectedAPI.Get("/versions/:id/release-notes", h.GetVersionReleaseNotes)

	// Calendar (EWS) - specific routes MUST come before parametric routes
	protectedAPI.Post("/calendar/sync", h.SyncCalendar)
	protectedAPI.Post("/calendar/free-slots", h.FindFreeSlots)
	protectedAPI.Post("/calendar/create", h.CreateCalendarEvent)
	protectedAPI.Get("/calendar/free-slots/simple", h.FreeSlotsSimple)
	protectedAPI.Get("/calendar/rooms", h.GetMeetingRooms)
	protectedAPI.Get("/calendar/:id", h.GetCalendar)
	protectedAPI.Get("/calendar/:id/simple", h.GetCalendarSimple)

	// Analytics
	protectedAPI.Get("/analytics/dashboard", h.GetDashboard)
	protectedAPI.Get("/analytics/employee/:id", h.GetEmployeeAnalytics)
	protectedAPI.Get("/analytics/employee/:id/by-category", h.GetEmployeeAnalyticsByCategory)
	protectedAPI.Get("/analytics/team/:id", h.GetTeamStats)

	// AD Integration
	protectedAPI.Post("/ad/sync", h.SyncADUsers)
	protectedAPI.Get("/ad/subordinates/:id", h.GetSubordinates)

	// JWT Token Management
	protectedAPI.Post("/auth/refresh", h.RefreshToken)
	protectedAPI.Get("/auth/me", h.GetMe)

	// Speech-to-Text
	protectedAPI.Post("/speech/transcribe", h.TranscribeAudio)

	// File Storage
	protectedAPI.Post("/files", h.UploadFile)
	protectedAPI.Get("/files", h.ListFiles)
	protectedAPI.Get("/files/:id", h.GetFile)
	protectedAPI.Get("/files/:id/url", h.GetFileURL)
	protectedAPI.Delete("/files/:id", h.DeleteFile)
	protectedAPI.Post("/files/attach", h.AttachFileToEntity)

	// BPMN / Camunda
	protectedAPI.Get("/bpmn/status", h.BPMNStatus)
	protectedAPI.Get("/bpmn/definitions", h.ListProcessDefinitions)
	protectedAPI.Get("/bpmn/definitions/:key", h.GetProcessDefinition)
	protectedAPI.Post("/bpmn/processes", h.StartProcess)
	protectedAPI.Get("/bpmn/processes", h.ListProcessInstances)
	protectedAPI.Get("/bpmn/processes/:id", h.GetProcessInstance)
	protectedAPI.Delete("/bpmn/processes/:id", h.DeleteProcessInstance)
	protectedAPI.Get("/bpmn/tasks", h.ListBPMNTasks)
	protectedAPI.Get("/bpmn/tasks/:id", h.GetBPMNTask)
	protectedAPI.Post("/bpmn/tasks/:id/complete", h.CompleteBPMNTask)
	protectedAPI.Post("/bpmn/tasks/:id/claim", h.ClaimBPMNTask)
	protectedAPI.Post("/bpmn/tasks/:id/unclaim", h.UnclaimBPMNTask)

	// Messenger
	protectedAPI.Get("/conversations", h.ListConversations)
	protectedAPI.Post("/conversations", h.CreateConversation)
	protectedAPI.Get("/conversations/:id", h.GetConversation)
	protectedAPI.Post("/messages", h.SendMessage)
	protectedAPI.Put("/messages/:id", h.UpdateMessage)
	protectedAPI.Delete("/messages/:id", h.DeleteMessage)
	protectedAPI.Post("/messages/:id/reactions", h.AddReaction)
	protectedAPI.Get("/messages/:id/reactions", h.GetReactions)

	// Telegram integration for channels (protected)
	protectedAPI.Get("/channels/:channel_id/telegram", h.GetTelegramConfig)
	protectedAPI.Post("/channels/:channel_id/telegram", h.ConfigureTelegramBot)

	// Telegram webhook (public - called by Telegram)
	api.Post("/telegram/webhook/:channel_id", h.TelegramWebhook)

	// GIPHY - GIF search (public API, no auth needed)
	api.Get("/gifs/search", h.SearchGifs)
	api.Get("/gifs/trending", h.TrendingGifs)

	// Mail (EWS)
	protectedAPI.Get("/mail/folders", h.GetMailFolders)
	protectedAPI.Get("/mail/emails", h.GetEmails)
	protectedAPI.Post("/mail/body", h.GetEmailBody)
	protectedAPI.Post("/mail/send", h.SendEmail)
	protectedAPI.Post("/mail/mark-read", h.MarkEmailAsRead)
	protectedAPI.Delete("/mail/email", h.DeleteEmail)
	protectedAPI.Post("/mail/attachments", h.GetAttachments)
	protectedAPI.Post("/mail/attachment/content", h.GetAttachmentContent)
	protectedAPI.Post("/mail/meeting/respond", h.RespondToMeeting)

	// Confluence
	protectedAPI.Get("/confluence/status", h.ConfluenceStatus)
	protectedAPI.Get("/confluence/spaces", h.GetConfluenceSpaces)
	protectedAPI.Get("/confluence/spaces/:key", h.GetConfluenceSpace)
	protectedAPI.Get("/confluence/spaces/:key/content", h.GetConfluenceSpaceContent)
	protectedAPI.Get("/confluence/pages/:id", h.GetConfluencePage)
	protectedAPI.Get("/confluence/pages/:id/children", h.GetConfluenceChildPages)
	protectedAPI.Get("/confluence/search", h.SearchConfluence)
	protectedAPI.Get("/confluence/recent", h.GetRecentConfluencePages)

	// GitHub Integration
	protectedAPI.Get("/github/status", h.GetGitHubStatus)
	protectedAPI.Post("/github/parse-url", h.ParseRepoURL)
	protectedAPI.Get("/github/repos/:owner/:repo", h.GetRepository)
	protectedAPI.Get("/github/repos/:owner/:repo/commits", h.GetCommits)
	protectedAPI.Get("/github/repos/:owner/:repo/branches", h.GetBranches)
	protectedAPI.Get("/github/repos/:owner/:repo/pulls", h.GetPullRequests)
	protectedAPI.Get("/github/tasks/:id/commits", h.GetTaskCommits)

	// User role (for current user)
	protectedAPI.Get("/auth/role", h.GetCurrentUserRole)

	// Admin routes (requires admin role)
	adminAPI := api.Group("/admin", middleware.JWTAuth(h.JWT), middleware.AdminAuth(h.DB))
	adminAPI.Use(middleware.CSRFProtection())
	adminAPI.Get("/stats", h.GetAdminStats)
	adminAPI.Get("/users", h.ListUsers)
	adminAPI.Put("/users/:id/role", h.UpdateUserRole)
	adminAPI.Get("/settings", h.GetSystemSettings)
	adminAPI.Put("/settings", h.UpdateSystemSetting)
	adminAPI.Get("/audit-logs", h.GetAuditLogs)
	adminAPI.Get("/departments", h.GetDepartments)

	// WebSocket for messenger
	app.Use("/ws/messenger", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/messenger", websocket.New(h.MessengerWebSocket))

	// WebSocket for connector
	app.Use("/ws/connector", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/connector", websocket.New(h.ConnectorWebSocket))

	// Start server
	log := utils.GetLogger()
	log.Info("Starting server", map[string]interface{}{
		"port":           cfg.Port,
		"ews_configured": cfg.EWSURL != "",
		"ad_configured":  cfg.ADURL != "",
	})

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Error("Server failed to start", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
}
