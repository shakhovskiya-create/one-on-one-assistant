package main

import (
	"log"
	"os"
	"time"

	"github.com/ekf/one-on-one-backend/internal/config"
	"github.com/ekf/one-on-one-backend/internal/handlers"
	"github.com/ekf/one-on-one-backend/internal/middleware"
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
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	// Health endpoints
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "EKF Hub API",
			"version": "1.0.0",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":       "healthy",
			"database":     h.DB != nil,
			"ews_calendar": cfg.EWSURL != "",
			"ews_url":      cfg.EWSURL,
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Public routes (no JWT required)
	publicAPI := api.Group("")
	publicAPI.Post("/ad/authenticate", h.AuthenticateAD)
	publicAPI.Get("/connector/status", h.ConnectorStatus)

	// Protected routes (JWT required)
	protectedAPI := api.Group("", middleware.JWTAuth(h.JWT))

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
	protectedAPI.Post("/process-meeting", h.ProcessMeeting)

	// Tasks
	protectedAPI.Get("/tasks", h.ListTasks)
	protectedAPI.Post("/tasks", h.CreateTask)
	protectedAPI.Get("/tasks/:id", h.GetTask)
	protectedAPI.Put("/tasks/:id", h.UpdateTask)
	protectedAPI.Delete("/tasks/:id", h.DeleteTask)

	// Kanban
	protectedAPI.Get("/kanban", h.GetKanban)
	protectedAPI.Put("/kanban/move", h.MoveTaskKanban)

	// Calendar (EWS) - specific routes MUST come before parametric routes
	protectedAPI.Post("/calendar/sync", h.SyncCalendar)
	protectedAPI.Post("/calendar/free-slots", h.FindFreeSlots)
	protectedAPI.Get("/calendar/free-slots/simple", h.FreeSlotsSimple)
	protectedAPI.Get("/calendar/:id", h.GetCalendar)
	protectedAPI.Get("/calendar/:id/simple", h.GetCalendarSimple)

	// Analytics
	protectedAPI.Get("/analytics/dashboard", h.GetDashboard)
	protectedAPI.Get("/analytics/employee/:id", h.GetEmployeeAnalytics)
	protectedAPI.Get("/analytics/employee/:id/by-category", h.GetEmployeeAnalyticsByCategory)

	// AD Integration
	protectedAPI.Post("/ad/sync", h.SyncADUsers)
	protectedAPI.Get("/ad/subordinates/:id", h.GetSubordinates)

	// JWT Token Management
	protectedAPI.Post("/auth/refresh", h.RefreshToken)
	protectedAPI.Get("/auth/me", h.GetMe)

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

	// Mail (EWS)
	protectedAPI.Get("/mail/folders", h.GetMailFolders)
	protectedAPI.Get("/mail/emails", h.GetEmails)
	protectedAPI.Post("/mail/body", h.GetEmailBody)
	protectedAPI.Post("/mail/send", h.SendEmail)
	protectedAPI.Post("/mail/mark-read", h.MarkEmailAsRead)
	protectedAPI.Delete("/mail/email", h.DeleteEmail)

	// Legacy routes (for backward compatibility)
	app.Get("/employees", h.ListEmployees)
	app.Post("/employees", h.CreateEmployee)
	app.Get("/employees/:id", h.GetEmployee)
	app.Put("/employees/:id", h.UpdateEmployee)
	app.Delete("/employees/:id", h.DeleteEmployee)
	app.Get("/employees/:id/dossier", h.GetEmployeeDossier)

	app.Get("/projects", h.ListProjects)
	app.Post("/projects", h.CreateProject)
	app.Get("/projects/:id", h.GetProject)
	app.Put("/projects/:id", h.UpdateProject)
	app.Delete("/projects/:id", h.DeleteProject)

	app.Get("/meetings", h.ListMeetings)
	app.Get("/meetings/:id", h.GetMeeting)
	app.Get("/meeting-categories", h.ListMeetingCategories)
	app.Post("/process-meeting", h.ProcessMeeting)

	app.Get("/tasks", h.ListTasks)
	app.Post("/tasks", h.CreateTask)
	app.Get("/tasks/:id", h.GetTask)
	app.Put("/tasks/:id", h.UpdateTask)
	app.Delete("/tasks/:id", h.DeleteTask)

	app.Get("/kanban", h.GetKanban)
	app.Put("/kanban/move", h.MoveTaskKanban)

	app.Get("/analytics/dashboard", h.GetDashboard)
	app.Get("/analytics/employee/:id", h.GetEmployeeAnalytics)
	app.Get("/analytics/employee/:id/by-category", h.GetEmployeeAnalyticsByCategory)

	app.Get("/connector/status", h.ConnectorStatus)
	app.Post("/ad/sync", h.SyncADUsers)
	app.Post("/ad/sync-direct", h.SyncADUsersDirect)
	app.Post("/ad/authenticate", h.AuthenticateAD)

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
	log.Printf("Starting server on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
