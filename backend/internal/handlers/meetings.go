package handlers

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/ekf/one-on-one-backend/pkg/ai"
	"github.com/gofiber/fiber/v2"
)

// ListMeetings returns meetings with filters
func (h *Handler) ListMeetings(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("meetings").Select("*, employees(name, position), projects(name), meeting_categories(code, name)")

	if employeeID := c.Query("employee_id"); employeeID != "" {
		query = query.Eq("employee_id", employeeID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Eq("project_id", projectID)
	}

	limit := c.QueryInt("limit", 50)

	var meetings []models.Meeting
	err := query.Order("date", true).Limit(limit).Execute(&meetings)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(meetings)
}

// GetMeeting returns a single meeting with details
func (h *Handler) GetMeeting(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var meeting models.Meeting
	err := h.DB.From("meetings").
		Select("*, employees(name, position), projects(name), meeting_categories(code, name)").
		Eq("id", id).Single().Execute(&meeting)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Meeting not found"})
	}

	// Get participants
	var participants []models.MeetingParticipant
	h.DB.From("meeting_participants").
		Select("*, employees(id, name, position)").
		Eq("meeting_id", id).Execute(&participants)

	// Get agreements
	var agreements []models.Agreement
	h.DB.From("agreements").Select("*").Eq("meeting_id", id).Execute(&agreements)

	return c.JSON(fiber.Map{
		"id":                 meeting.ID,
		"title":              meeting.Title,
		"employee_id":        meeting.EmployeeID,
		"project_id":         meeting.ProjectID,
		"category_id":        meeting.CategoryID,
		"date":               meeting.Date,
		"start_time":         meeting.StartTime,
		"end_time":           meeting.EndTime,
		"location":           meeting.Location,
		"transcript":         meeting.Transcript,
		"summary":            meeting.Summary,
		"mood_score":         meeting.MoodScore,
		"analysis":           meeting.Analysis,
		"exchange_id":        meeting.ExchangeID,
		"exchange_data":      meeting.ExchangeData,
		"employees":          meeting.Employee,
		"projects":           meeting.Project,
		"meeting_categories": meeting.Category,
		"participants":       extractEmployees(participants),
		"agreements":         agreements,
	})
}

// CreateMeeting creates a new meeting
func (h *Handler) CreateMeeting(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var req struct {
		Title      string  `json:"title"`
		EmployeeID string  `json:"employee_id"`
		ProjectID  *string `json:"project_id"`
		CategoryID *string `json:"category_id"`
		Date       string  `json:"date"`
		StartTime  *string `json:"start_time"`
		EndTime    *string `json:"end_time"`
		Location   *string `json:"location"`
		Summary    *string `json:"summary"`
		MoodScore  *int    `json:"mood_score"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title is required"})
	}
	if req.EmployeeID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Employee ID is required"})
	}
	if req.Date == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Date is required"})
	}

	// Create meeting
	meeting := map[string]interface{}{
		"title":       req.Title,
		"employee_id": req.EmployeeID,
		"date":        req.Date,
	}

	if req.ProjectID != nil {
		meeting["project_id"] = *req.ProjectID
	}
	if req.CategoryID != nil {
		meeting["category_id"] = *req.CategoryID
	}
	if req.StartTime != nil {
		meeting["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		meeting["end_time"] = *req.EndTime
	}
	if req.Location != nil {
		meeting["location"] = *req.Location
	}
	if req.Summary != nil {
		meeting["summary"] = *req.Summary
	}
	if req.MoodScore != nil {
		meeting["mood_score"] = *req.MoodScore
	}

	result, err := h.DB.Insert("meetings", meeting)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create meeting", "details": err.Error()})
	}

	var created []models.Meeting
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create meeting"})
	}

	return c.Status(201).JSON(created[0])
}

// ListMeetingCategories returns all meeting categories
func (h *Handler) ListMeetingCategories(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var categories []models.MeetingCategory
	err := h.DB.From("meeting_categories").Select("*").Order("name", false).Execute(&categories)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(categories)
}

// AIStatus returns status of AI transcription services
func (h *Handler) AIStatus(c *fiber.Ctx) error {
	status := fiber.Map{
		"transcription": fiber.Map{
			"whisper": h.AI != nil && h.Config.OpenAIKey != "",
			"yandex":  h.AI != nil && h.Config.YandexAPIKey != "" && h.Config.YandexFolderID != "",
		},
		"analysis": fiber.Map{
			"openai":    h.AI != nil && h.Config.OpenAIKey != "",
			"anthropic": h.AI != nil && h.Config.AnthropicKey != "",
		},
		"available": h.AI != nil && (h.Config.OpenAIKey != "" || (h.Config.YandexAPIKey != "" && h.Config.YandexFolderID != "")),
	}

	if !status["available"].(bool) {
		status["hint"] = "Для работы транскрипции необходимо установить OPENAI_API_KEY или YANDEX_API_KEY + YANDEX_FOLDER_ID"
	}

	return c.JSON(status)
}

// ProcessMeeting processes an audio file and creates a meeting
func (h *Handler) ProcessMeeting(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	// Parse form
	categoryCode := c.FormValue("category_code", "one_on_one")
	employeeID := c.FormValue("employee_id")
	projectID := c.FormValue("project_id")
	meetingDate := c.FormValue("meeting_date")
	title := c.FormValue("title")
	participantIDsJSON := c.FormValue("participant_ids", "[]")

	var participantIDs []string
	json.Unmarshal([]byte(participantIDsJSON), &participantIDs)

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Audio file required"})
	}

	// Save to temp file
	tmpFile, err := os.CreateTemp("", "audio-*"+filepath.Ext(file.Filename))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create temp file"})
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to open uploaded file"})
	}
	defer src.Close()

	if _, err := io.Copy(tmpFile, src); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	}
	tmpFile.Close()

	// Transcribe with both services
	whisperTranscript, _ := h.AI.TranscribeWhisper(tmpFile.Name())
	yandexTranscript, _ := h.AI.TranscribeYandex(tmpFile.Name())

	// Merge transcripts
	var mergedTranscript string
	if whisperTranscript != "" && yandexTranscript != "" {
		mergedTranscript, _ = h.AI.MergeTranscripts(whisperTranscript, yandexTranscript)
	} else if whisperTranscript != "" {
		mergedTranscript = whisperTranscript
	} else {
		mergedTranscript = yandexTranscript
	}

	if mergedTranscript == "" {
		// Provide more helpful error message
		aiStatus := "Сервисы транскрипции недоступны. "
		if h.AI == nil {
			aiStatus += "AI клиент не инициализирован. "
		} else {
			aiStatus += "Проверьте настройки: OPENAI_API_KEY или YANDEX_API_KEY. "
		}
		return c.Status(500).JSON(fiber.Map{
			"error":   "Transcription failed",
			"details": aiStatus,
			"hint":    "Установите OPENAI_API_KEY в переменных окружения для работы транскрипции",
		})
	}

	// Build context
	ctx := ai.AnalysisContext{
		Transcript: mergedTranscript,
	}

	if employeeID != "" {
		ctx.EmployeeContext = h.getEmployeeContext(employeeID)
		ctx.MeetingsHistory = h.getEmployeeMeetingsHistory(employeeID)
	}
	if projectID != "" {
		ctx.ProjectContext = h.getProjectContext(projectID)
		if ctx.MeetingsHistory == "" {
			ctx.MeetingsHistory = h.getProjectMeetingsHistory(projectID)
		}
	}
	if len(participantIDs) > 0 {
		ctx.Participants = h.getParticipantsInfo(participantIDs)
	}

	// Analyze
	analysis, err := h.AI.Analyze(categoryCode, ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Analysis failed: " + err.Error()})
	}

	// Get category ID
	var category models.MeetingCategory
	h.DB.From("meeting_categories").Select("id").Eq("code", categoryCode).Single().Execute(&category)

	// Create meeting
	meetingData := map[string]interface{}{
		"title":              title,
		"employee_id":        nilIfEmpty(employeeID),
		"project_id":         nilIfEmpty(projectID),
		"category_id":        nilIfEmpty(category.ID),
		"date":               meetingDate,
		"transcript_whisper": whisperTranscript,
		"transcript_yandex":  yandexTranscript,
		"transcript_merged":  mergedTranscript,
		"transcript":         mergedTranscript,
		"summary":            analysis["summary"],
		"mood_score":         analysis["mood_score"],
		"analysis":           analysis,
	}

	if title == "" {
		meetingData["title"] = categoryCode + " - " + meetingDate
	}

	result, err := h.DB.Insert("meetings", meetingData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []map[string]interface{}
	json.Unmarshal(result, &created)

	var meetingID string
	if len(created) > 0 {
		meetingID, _ = created[0]["id"].(string)
	}

	// Add participants
	for _, pid := range participantIDs {
		h.DB.Insert("meeting_participants", map[string]interface{}{
			"meeting_id":  meetingID,
			"employee_id": pid,
		})
	}

	// Save agreements/action items
	actionItems := getActionItems(analysis)
	for _, item := range actionItems {
		task, _ := item["task"].(string)
		if task == "" {
			task, _ = item["improvement"].(string)
		}
		if task == "" {
			continue
		}

		agreementData := map[string]interface{}{
			"meeting_id":  meetingID,
			"task":        task,
			"responsible": item["responsible"],
			"deadline":    item["deadline"],
			"status":      "pending",
		}
		h.DB.Insert("agreements", agreementData)

		// Also create task
		taskData := map[string]interface{}{
			"title":       task,
			"description": "Из встречи: " + meetingDate,
			"status":      "todo",
			"priority":    3,
			"meeting_id":  meetingID,
			"project_id":  nilIfEmpty(projectID),
			"due_date":    item["deadline"],
		}
		if employeeID != "" {
			taskData["assignee_id"] = employeeID
		}
		h.DB.Insert("tasks", taskData)
	}

	return c.JSON(fiber.Map{
		"meeting_id": meetingID,
		"transcript": fiber.Map{
			"whisper": truncate(whisperTranscript, 500),
			"yandex":  truncate(yandexTranscript, 500),
			"merged":  mergedTranscript,
		},
		"analysis": analysis,
	})
}

func (h *Handler) getEmployeeContext(employeeID string) string {
	var employee models.Employee
	h.DB.From("employees").Select("*").Eq("id", employeeID).Single().Execute(&employee)

	var tasks []models.Task
	h.DB.From("tasks").Select("status, due_date").Eq("assignee_id", employeeID).Execute(&tasks)

	total := len(tasks)
	done := 0
	inProgress := 0
	overdue := 0
	today := getCurrentDate()

	for _, t := range tasks {
		switch t.Status {
		case "done":
			done++
		case "in_progress":
			inProgress++
		}
		if t.DueDate != nil && *t.DueDate < today && t.Status != "done" {
			overdue++
		}
	}

	return "ИМЯ: " + employee.Name + "\n" +
		"ДОЛЖНОСТЬ: " + employee.Position + "\n" +
		"СТАТИСТИКА ЗАДАЧ: всего " + itoa(total) + ", выполнено " + itoa(done) + ", в работе " + itoa(inProgress) + ", просрочено " + itoa(overdue)
}

func (h *Handler) getEmployeeMeetingsHistory(employeeID string) string {
	var meetings []models.Meeting
	h.DB.From("meetings").Select("date, summary, mood_score").Eq("employee_id", employeeID).Order("date", true).Limit(5).Execute(&meetings)

	var history []string
	for _, m := range meetings {
		mood := "?"
		if m.MoodScore != nil {
			mood = itoa(*m.MoodScore)
		}
		summary := ""
		if m.Summary != nil {
			summary = truncate(*m.Summary, 200)
		}
		history = append(history, "[1-на-1 "+m.Date+"] Настроение: "+mood+"/10\n"+summary)
	}

	if len(history) == 0 {
		return "Предыдущих встреч не найдено"
	}
	return strings.Join(history, "\n\n")
}

func (h *Handler) getProjectContext(projectID string) string {
	var project models.Project
	h.DB.From("projects").Select("*").Eq("id", projectID).Single().Execute(&project)

	var tasks []struct {
		Status string `json:"status"`
	}
	h.DB.From("tasks").Select("status").Eq("project_id", projectID).Execute(&tasks)

	total := len(tasks)
	done := 0
	for _, t := range tasks {
		if t.Status == "done" {
			done++
		}
	}
	progress := 0
	if total > 0 {
		progress = (done * 100) / total
	}

	desc := ""
	if project.Description != nil {
		desc = *project.Description
	}

	return "ПРОЕКТ: " + project.Name + "\n" +
		"ОПИСАНИЕ: " + desc + "\n" +
		"СТАТУС: " + project.Status + "\n" +
		"ПРОГРЕСС: " + itoa(progress) + "% (" + itoa(done) + "/" + itoa(total) + " задач)"
}

func (h *Handler) getProjectMeetingsHistory(projectID string) string {
	var meetings []models.Meeting
	h.DB.From("meetings").Select("date, title, summary").Eq("project_id", projectID).Order("date", true).Limit(5).Execute(&meetings)

	var history []string
	for _, m := range meetings {
		title := ""
		if m.Title != nil {
			title = *m.Title
		}
		summary := ""
		if m.Summary != nil {
			summary = truncate(*m.Summary, 150)
		}
		history = append(history, "["+m.Date+"] "+title+"\n"+summary)
	}

	if len(history) == 0 {
		return ""
	}
	return strings.Join(history, "\n\n")
}

func (h *Handler) getParticipantsInfo(ids []string) string {
	var employees []models.Employee
	h.DB.From("employees").Select("name, position").In("id", ids).Execute(&employees)

	var lines []string
	for _, e := range employees {
		lines = append(lines, "- "+e.Name+" ("+e.Position+")")
	}
	return strings.Join(lines, "\n")
}

func extractEmployees(participants []models.MeetingParticipant) []models.Employee {
	var employees []models.Employee
	for _, p := range participants {
		if p.Employee != nil {
			employees = append(employees, *p.Employee)
		}
	}
	return employees
}

func getActionItems(analysis map[string]interface{}) []map[string]interface{} {
	var items []map[string]interface{}

	if agreements, ok := analysis["agreements"].([]interface{}); ok {
		for _, a := range agreements {
			if item, ok := a.(map[string]interface{}); ok {
				items = append(items, item)
			}
		}
	}

	if actionItems, ok := analysis["action_items"].([]interface{}); ok {
		for _, a := range actionItems {
			if item, ok := a.(map[string]interface{}); ok {
				items = append(items, item)
			}
		}
	}

	return items
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func itoa(n int) string {
	// Use standard library for integer to string conversion
	return strconv.Itoa(n)
}

func getCurrentDate() string {
	// Return current date in ISO 8601 format
	return time.Now().Format("2006-01-02")
}
