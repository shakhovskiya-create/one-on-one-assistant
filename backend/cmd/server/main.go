package main

import (
	"log"
	"os"

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
	if cfg.EWSSkipTLSVerify && os.Getenv("ENV") == "production" {
		log.Fatal("EWS_SKIP_TLS_VERIFY cannot be enabled in production")
	}

	// Create handler
	h := handlers.NewHandler(cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB for audio files
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "*",
	}))

	// Health endpoints
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "EKF Team Hub API",
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
	protected := api.Group("", middleware.JWTAuth(h.JWT))

	// Employees
	protected.Get("/employees", h.ListEmployees)
	protected.Post("/employees", h.CreateEmployee)
	protected.Get("/employees/:id", h.GetEmployee)
	protected.Put("/employees/:id", h.UpdateEmployee)
	protected.Delete("/employees/:id", h.DeleteEmployee)
	protected.Get("/employees/:id/dossier", h.GetEmployeeDossier)
	protected.Get("/employees/my-team", h.GetMyTeam)

	// Projects
	protected.Get("/projects", h.ListProjects)
	protected.Post("/projects", h.CreateProject)
	protected.Get("/projects/:id", h.GetProject)
	protected.Put("/projects/:id", h.UpdateProject)
	protected.Delete("/projects/:id", h.DeleteProject)

	// Meetings
	protected.Get("/meetings", h.ListMeetings)
	protected.Post("/meetings", h.CreateMeeting)
	protected.Get("/meetings/:id", h.GetMeeting)
	protected.Get("/meeting-categories", h.ListMeetingCategories)
	protected.Post("/process-meeting", h.ProcessMeeting)

	// Tasks
	protected.Get("/tasks", h.ListTasks)
	protected.Post("/tasks", h.CreateTask)
	protected.Get("/tasks/:id", h.GetTask)
	protected.Put("/tasks/:id", h.UpdateTask)
	protected.Delete("/tasks/:id", h.DeleteTask)
	protected.Post("/tasks/:id/comments", h.AddTaskComment)

	// Kanban
	protected.Get("/kanban", h.GetKanban)
	protected.Put("/kanban/move", h.MoveTaskKanban)

	// Calendar (EWS) - specific routes MUST come before parametric routes
	protected.Post("/calendar/sync", h.SyncCalendar)
	protected.Post("/calendar/free-slots", h.FindFreeSlots)
	protected.Get("/calendar/free-slots/simple", h.FreeSlotsSimple)
	protected.Post("/calendar/:id", h.GetCalendar)
	protected.Get("/calendar/:id/simple", h.GetCalendarSimple)

	// Analytics
	protected.Get("/analytics/dashboard", h.GetDashboard)
	protected.Get("/analytics/employee/:id", h.GetEmployeeAnalytics)
	protected.Get("/analytics/employee/:id/by-category", h.GetEmployeeAnalyticsByCategory)

	// AD Integration
	protected.Get("/connector/status", h.ConnectorStatus)
	protected.Post("/ad/sync", h.SyncADUsers)
	api.Post("/ad/authenticate", h.AuthenticateAD)
	protected.Get("/ad/subordinates/:id", h.GetSubordinates)

	// Local User Management (works without connector)
	protected.Post("/users/local", h.CreateLocalUser)
	protected.Post("/users/set-password", h.SetLocalPassword)
	protected.Post("/users/change-password", h.ChangePassword)

	// JWT Token Management
	api.Post("/auth/refresh", h.RefreshToken)
	protected.Get("/auth/me", h.GetMe)

	// File Storage
	protected.Post("/files", h.UploadFile)
	protected.Get("/files", h.ListFiles)
	protected.Get("/files/:id", h.GetFile)
	protected.Get("/files/:id/url", h.GetFileURL)
	protected.Delete("/files/:id", h.DeleteFile)
	protected.Post("/files/attach", h.AttachFileToEntity)

	// BPMN / Camunda
	protected.Get("/bpmn/status", h.BPMNStatus)
	protected.Get("/bpmn/definitions", h.ListProcessDefinitions)
	protected.Get("/bpmn/definitions/:key", h.GetProcessDefinition)
	protected.Post("/bpmn/processes", h.StartProcess)
	protected.Get("/bpmn/processes", h.ListProcessInstances)
	protected.Get("/bpmn/processes/:id", h.GetProcessInstance)
	protected.Delete("/bpmn/processes/:id", h.DeleteProcessInstance)
	protected.Get("/bpmn/tasks", h.ListBPMNTasks)
	protected.Get("/bpmn/tasks/:id", h.GetBPMNTask)
	protected.Post("/bpmn/tasks/:id/complete", h.CompleteBPMNTask)
	protected.Post("/bpmn/tasks/:id/claim", h.ClaimBPMNTask)
	protected.Post("/bpmn/tasks/:id/unclaim", h.UnclaimBPMNTask)

	// Messenger
	protected.Get("/conversations", h.ListConversations)
	protected.Post("/conversations", h.CreateConversation)
	protected.Get("/conversations/:id", h.GetConversation)
	protected.Post("/messages", h.SendMessage)

	// Legacy routes (for backward compatibility)
	legacy := app.Group("", middleware.JWTAuth(h.JWT))
	legacy.Get("/employees", h.ListEmployees)
	legacy.Post("/employees", h.CreateEmployee)
	legacy.Get("/employees/:id", h.GetEmployee)
	legacy.Put("/employees/:id", h.UpdateEmployee)
	legacy.Delete("/employees/:id", h.DeleteEmployee)
	legacy.Get("/employees/:id/dossier", h.GetEmployeeDossier)

	legacy.Get("/projects", h.ListProjects)
	legacy.Post("/projects", h.CreateProject)
	legacy.Get("/projects/:id", h.GetProject)
	legacy.Put("/projects/:id", h.UpdateProject)
	legacy.Delete("/projects/:id", h.DeleteProject)

	legacy.Get("/meetings", h.ListMeetings)
	legacy.Post("/meetings", h.CreateMeeting)
	legacy.Get("/meetings/:id", h.GetMeeting)
	legacy.Get("/meeting-categories", h.ListMeetingCategories)
	legacy.Post("/process-meeting", h.ProcessMeeting)

	legacy.Get("/tasks", h.ListTasks)
	legacy.Post("/tasks", h.CreateTask)
	legacy.Get("/tasks/:id", h.GetTask)
	legacy.Put("/tasks/:id", h.UpdateTask)
	legacy.Delete("/tasks/:id", h.DeleteTask)
	legacy.Post("/tasks/:id/comments", h.AddTaskComment)

	legacy.Get("/kanban", h.GetKanban)
	legacy.Put("/kanban/move", h.MoveTaskKanban)

	legacy.Get("/analytics/dashboard", h.GetDashboard)
	legacy.Get("/analytics/employee/:id", h.GetEmployeeAnalytics)
	legacy.Get("/analytics/employee/:id/by-category", h.GetEmployeeAnalyticsByCategory)

	legacy.Get("/connector/status", h.ConnectorStatus)
	legacy.Post("/ad/sync", h.SyncADUsers)
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
