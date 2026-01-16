package main

import (
	"log"

	"github.com/ekf/one-on-one-backend/internal/config"
	"github.com/ekf/one-on-one-backend/internal/handlers"
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

	// Employees
	api.Get("/employees", h.ListEmployees)
	api.Post("/employees", h.CreateEmployee)
	api.Get("/employees/:id", h.GetEmployee)
	api.Put("/employees/:id", h.UpdateEmployee)
	api.Delete("/employees/:id", h.DeleteEmployee)
	api.Get("/employees/:id/dossier", h.GetEmployeeDossier)
	api.Get("/employees/my-team", h.GetMyTeam)

	// Projects
	api.Get("/projects", h.ListProjects)
	api.Post("/projects", h.CreateProject)
	api.Get("/projects/:id", h.GetProject)
	api.Put("/projects/:id", h.UpdateProject)
	api.Delete("/projects/:id", h.DeleteProject)

	// Meetings
	api.Get("/meetings", h.ListMeetings)
	api.Get("/meetings/:id", h.GetMeeting)
	api.Get("/meeting-categories", h.ListMeetingCategories)
	api.Post("/process-meeting", h.ProcessMeeting)

	// Tasks
	api.Get("/tasks", h.ListTasks)
	api.Post("/tasks", h.CreateTask)
	api.Get("/tasks/:id", h.GetTask)
	api.Put("/tasks/:id", h.UpdateTask)
	api.Delete("/tasks/:id", h.DeleteTask)

	// Kanban
	api.Get("/kanban", h.GetKanban)
	api.Put("/kanban/move", h.MoveTaskKanban)

	// Calendar (EWS) - specific routes MUST come before parametric routes
	api.Post("/calendar/sync", h.SyncCalendar)
	api.Post("/calendar/free-slots", h.FindFreeSlots)
	api.Get("/calendar/free-slots/simple", h.FreeSlotsSimple)
	api.Post("/calendar/:id", h.GetCalendar)
	api.Get("/calendar/:id/simple", h.GetCalendarSimple)

	// Analytics
	api.Get("/analytics/dashboard", h.GetDashboard)
	api.Get("/analytics/employee/:id", h.GetEmployeeAnalytics)
	api.Get("/analytics/employee/:id/by-category", h.GetEmployeeAnalyticsByCategory)

	// AD Integration
	api.Get("/connector/status", h.ConnectorStatus)
	api.Post("/ad/sync", h.SyncADUsers)
	api.Post("/ad/authenticate", h.AuthenticateAD)
	api.Get("/ad/subordinates/:id", h.GetSubordinates)

	// Local User Management (works without connector)
	api.Post("/users/local", h.CreateLocalUser)
	api.Post("/users/set-password", h.SetLocalPassword)
	api.Post("/users/change-password", h.ChangePassword)

	// JWT Token Management
	api.Post("/auth/refresh", h.RefreshToken)
	api.Get("/auth/me", h.GetMe)

	// File Storage
	api.Post("/files", h.UploadFile)
	api.Get("/files", h.ListFiles)
	api.Get("/files/:id", h.GetFile)
	api.Get("/files/:id/url", h.GetFileURL)
	api.Delete("/files/:id", h.DeleteFile)
	api.Post("/files/attach", h.AttachFileToEntity)

	// BPMN / Camunda
	api.Get("/bpmn/status", h.BPMNStatus)
	api.Get("/bpmn/definitions", h.ListProcessDefinitions)
	api.Get("/bpmn/definitions/:key", h.GetProcessDefinition)
	api.Post("/bpmn/processes", h.StartProcess)
	api.Get("/bpmn/processes", h.ListProcessInstances)
	api.Get("/bpmn/processes/:id", h.GetProcessInstance)
	api.Delete("/bpmn/processes/:id", h.DeleteProcessInstance)
	api.Get("/bpmn/tasks", h.ListBPMNTasks)
	api.Get("/bpmn/tasks/:id", h.GetBPMNTask)
	api.Post("/bpmn/tasks/:id/complete", h.CompleteBPMNTask)
	api.Post("/bpmn/tasks/:id/claim", h.ClaimBPMNTask)
	api.Post("/bpmn/tasks/:id/unclaim", h.UnclaimBPMNTask)

	// Messenger
	api.Get("/conversations", h.ListConversations)
	api.Post("/conversations", h.CreateConversation)
	api.Get("/conversations/:id", h.GetConversation)
	api.Post("/messages", h.SendMessage)

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
