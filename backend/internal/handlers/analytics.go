package handlers

import (
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// GetDashboard returns dashboard data
func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	period := c.Query("period", "month")
	startDate := getPeriodStart(period)
	managerID := c.Query("manager_id")

	// Get employees - filter by manager if specified
	var employees []models.Employee
	if managerID != "" {
		// Get all subordinates recursively
		employees = h.getSubordinatesRecursive(managerID)
	} else {
		h.DB.From("employees").Select("*").Execute(&employees)
	}

	// Get active projects
	var projects []models.Project
	h.DB.From("projects").Select("*").Eq("status", "active").Execute(&projects)

	// Get recent meetings - filter by employees if manager specified
	var meetings []models.Meeting
	meetingQuery := h.DB.From("meetings").Select("*, employees(name), meeting_categories(name)")
	if managerID != "" && len(employees) > 0 {
		meetingQuery = meetingQuery.In("employee_id", getEmployeeIDs(employees))
	}
	meetingQuery.Order("date", true).Limit(50).Execute(&meetings)

	// Get task stats - filter by employees if manager specified
	var tasks []struct {
		Status  string  `json:"status"`
		DueDate *string `json:"due_date"`
	}
	taskQuery := h.DB.From("tasks").Select("status, due_date")
	if managerID != "" && len(employees) > 0 {
		taskQuery = taskQuery.In("assignee_id", getEmployeeIDs(employees))
	}
	taskQuery.Execute(&tasks)

	// Get agreement stats
	var agreements []struct {
		Status   string  `json:"status"`
		Deadline *string `json:"deadline"`
	}
	h.DB.From("agreements").Select("status, deadline").Execute(&agreements)

	today := time.Now().Format("2006-01-02")
	totalTasks := len(tasks)
	doneTasks := 0
	inProgressTasks := 0
	overdueTasks := 0

	for _, t := range tasks {
		if t.DueDate != nil && startDate != "" && *t.DueDate < startDate {
			continue
		}
		switch t.Status {
		case "done":
			doneTasks++
		case "in_progress":
			inProgressTasks++
		}
		if t.DueDate != nil && *t.DueDate < today && t.Status != "done" {
			overdueTasks++
		}
	}

	// Find red flags
	var redFlags []fiber.Map
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		if m.Analysis != nil {
			if flags, ok := m.Analysis["red_flags"].(map[string]interface{}); ok {
				burnout, _ := flags["burnout_signs"].(string)
				turnover, _ := flags["turnover_risk"].(string)
				if burnout != "" || turnover == "medium" || turnover == "high" {
					employeeName := ""
					if m.Employee != nil {
						employeeName = m.Employee.Name
					}
					redFlags = append(redFlags, fiber.Map{
						"employee": employeeName,
						"date":     m.Date,
						"flags":    flags,
					})
				}
			}
		}
	}

	// Calculate average mood
	var avgMood float64
	var moodCount int
	var moodTrend []fiber.Map
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		if m.MoodScore != nil {
			avgMood += float64(*m.MoodScore)
			moodCount++
			moodTrend = append(moodTrend, fiber.Map{
				"date":  m.Date,
				"score": float64(*m.MoodScore),
			})
		}
	}
	if moodCount > 0 {
		avgMood /= float64(moodCount)
	}

	// Get meetings this month
	startOfMonth := time.Now().Format("2006-01") + "-01"
	var meetingsThisMonth int
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		if m.Date >= startOfMonth {
			meetingsThisMonth++
		}
	}

	// Meetings by category
	meetingsByCategory := make(map[string]int)
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		if m.Category != nil {
			meetingsByCategory[m.Category.Code]++
		}
	}

	totalAgreements := 0
	completedAgreements := 0
	overdueAgreements := 0
	for _, a := range agreements {
		if a.Deadline != nil && startDate != "" && *a.Deadline < startDate {
			continue
		}
		totalAgreements++
		if a.Status == "completed" || a.Status == "done" {
			completedAgreements++
			continue
		}
		if a.Deadline != nil && *a.Deadline < today {
			overdueAgreements++
		}
	}

	// Employees needing attention (no meetings for a while)
	var employeesNeedingAttention []fiber.Map
	// TODO: calculate based on last meeting date

	// Ensure arrays are not nil
	if redFlags == nil {
		redFlags = []fiber.Map{}
	}
	if moodTrend == nil {
		moodTrend = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"employees":                   employees,
		"projects":                    projects,
		"recent_meetings":             meetings,
		"total_employees":             len(employees),
		"meetings_this_month":         meetingsThisMonth,
		"average_mood":                avgMood,
		"tasks_completed":             doneTasks,
		"tasks_todo":                  totalTasks - doneTasks - inProgressTasks,
		"tasks_in_progress":           inProgressTasks,
		"mood_trend":                  moodTrend,
		"employees_needing_attention": employeesNeedingAttention,
		"meetings_by_category":        meetingsByCategory,
		"agreements_total":            totalAgreements,
		"agreements_completed":        completedAgreements,
		"agreements_overdue":          overdueAgreements,
		"task_summary": fiber.Map{
			"total":       totalTasks,
			"done":        doneTasks,
			"in_progress": inProgressTasks,
			"overdue":     overdueTasks,
		},
		"red_flags": redFlags,
	})
}

// GetEmployeeAnalytics returns analytics for an employee
func (h *Handler) GetEmployeeAnalytics(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	period := c.Query("period", "month")
	startDate := getPeriodStart(period)

	employeeID := c.Params("id")

	// Get meetings
	var meetings []models.Meeting
	h.DB.From("meetings").Select("id, date, mood_score, analysis").Eq("employee_id", employeeID).Order("date", false).Execute(&meetings)

	filteredMeetingsCount := 0

	// Mood history
	var moodHistory []fiber.Map
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		filteredMeetingsCount++
		if m.MoodScore != nil {
			moodHistory = append(moodHistory, fiber.Map{
				"date":  m.Date,
				"score": *m.MoodScore,
			})
		}
	}

	// Red flags history
	var redFlagsHistory []fiber.Map
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		if m.Analysis != nil {
			if flags, ok := m.Analysis["red_flags"].(map[string]interface{}); ok {
				burnout, _ := flags["burnout_signs"].(string)
				turnover, _ := flags["turnover_risk"].(string)
				if burnout != "" || (turnover != "low" && turnover != "") {
					redFlagsHistory = append(redFlagsHistory, fiber.Map{
						"date":  m.Date,
						"flags": flags,
					})
				}
			}
		}
	}

	// Task stats
	var tasks []struct {
		Status string  `json:"status"`
		Due    *string `json:"due_date"`
	}
	h.DB.From("tasks").Select("status, due_date").Eq("assignee_id", employeeID).Execute(&tasks)

	totalTasks := 0
	doneTasks := 0
	inProgressTasks := 0
	for _, t := range tasks {
		if t.Due != nil && startDate != "" && *t.Due < startDate {
			continue
		}
		totalTasks++
		switch t.Status {
		case "done":
			doneTasks++
		case "in_progress":
			inProgressTasks++
		}
	}

	// Agreement stats
	var meetingIDs []string
	for _, m := range meetings {
		if startDate != "" && m.Date < startDate {
			continue
		}
		meetingIDs = append(meetingIDs, m.ID)
	}

	var agreements []struct {
		Status   string  `json:"status"`
		Deadline *string `json:"deadline"`
	}
	if len(meetingIDs) > 0 {
		h.DB.From("agreements").Select("status, deadline").In("meeting_id", meetingIDs).Execute(&agreements)
	}

	today := time.Now().Format("2006-01-02")
	totalAgreements := len(agreements)
	completedAgreements := 0
	pendingAgreements := 0
	overdueAgreements := 0

	for _, a := range agreements {
		if a.Status == "completed" || a.Status == "done" {
			completedAgreements++
		} else if a.Status == "pending" {
			pendingAgreements++
			if a.Deadline != nil && *a.Deadline < today {
				overdueAgreements++
			}
		}
	}

	// Ensure arrays are not nil
	if moodHistory == nil {
		moodHistory = []fiber.Map{}
	}
	if redFlagsHistory == nil {
		redFlagsHistory = []fiber.Map{}
	}

	return c.JSON(fiber.Map{
		"mood_history":      moodHistory,
		"red_flags_history": redFlagsHistory,
		"task_stats": fiber.Map{
			"total":       totalTasks,
			"done":        doneTasks,
			"in_progress": inProgressTasks,
		},
		"agreement_stats": fiber.Map{
			"total":     totalAgreements,
			"completed": completedAgreements,
			"pending":   pendingAgreements - overdueAgreements,
			"overdue":   overdueAgreements,
		},
		"total_meetings": filteredMeetingsCount,
	})
}

func getPeriodStart(period string) string {
	now := time.Now()
	switch period {
	case "week":
		start := now.AddDate(0, 0, -7)
		return start.Format("2006-01-02")
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start.Format("2006-01-02")
	case "quarter":
		quarter := (now.Month()-1)/3 + 1
		startMonth := time.Month((quarter-1)*3 + 1)
		start := time.Date(now.Year(), startMonth, 1, 0, 0, 0, 0, now.Location())
		return start.Format("2006-01-02")
	case "year":
		start := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		return start.Format("2006-01-02")
	default:
		return ""
	}
}

// getSubordinatesRecursive gets all subordinates recursively for a manager
func (h *Handler) getSubordinatesRecursive(managerID string) []models.Employee {
	var result []models.Employee

	// Get direct subordinates
	var directSubordinates []models.Employee
	h.DB.From("employees").Select("*").Eq("manager_id", managerID).Execute(&directSubordinates)

	for _, sub := range directSubordinates {
		result = append(result, sub)
		// Recursively get their subordinates
		subOfSub := h.getSubordinatesRecursive(sub.ID)
		result = append(result, subOfSub...)
	}

	return result
}

// getEmployeeIDs extracts IDs from employee slice
func getEmployeeIDs(employees []models.Employee) []string {
	ids := make([]string, len(employees))
	for i, emp := range employees {
		ids[i] = emp.ID
	}
	return ids
}

// GetEmployeeAnalyticsByCategory returns analytics by meeting category
func (h *Handler) GetEmployeeAnalyticsByCategory(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	employeeID := c.Params("id")

	// Get categories
	var categories []models.MeetingCategory
	h.DB.From("meeting_categories").Select("*").Execute(&categories)

	// Get meetings
	var meetings []models.Meeting
	h.DB.From("meetings").Select("*, meeting_categories(code, name)").Eq("employee_id", employeeID).Execute(&meetings)

	var result []fiber.Map
	for _, cat := range categories {
		var catMeetings []models.Meeting
		for _, m := range meetings {
			if m.CategoryID != nil && *m.CategoryID == cat.ID {
				catMeetings = append(catMeetings, m)
			}
		}

		// Calculate avg mood
		var moodSum int
		var moodCount int
		var redFlagsCount int

		for _, m := range catMeetings {
			if m.MoodScore != nil {
				moodSum += *m.MoodScore
				moodCount++
			}
			if m.Analysis != nil {
				if flags, ok := m.Analysis["red_flags"].(map[string]interface{}); ok {
					if len(flags) > 0 {
						redFlagsCount++
					}
				}
			}
		}

		var avgMood interface{}
		if moodCount > 0 {
			avgMood = float64(moodSum) / float64(moodCount)
		}

		result = append(result, fiber.Map{
			"category":             cat,
			"meetings_count":       len(catMeetings),
			"avg_mood":             avgMood,
			"mood_trend":           nil,
			"agreements_created":   0,
			"agreements_completed": 0,
			"common_topics":        []string{},
			"red_flags_count":      redFlagsCount,
			"specific_metrics":     fiber.Map{},
		})
	}

	return c.JSON(result)
}

// GetTeamStats returns stats for each team member
func (h *Handler) GetTeamStats(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	managerID := c.Params("id")
	if managerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Manager ID required"})
	}

	// Get direct subordinates
	var subordinates []models.Employee
	h.DB.From("employees").Select("id, name, position, photo_base64").Eq("manager_id", managerID).Execute(&subordinates)

	today := time.Now().Format("2006-01-02")
	var result []fiber.Map

	for _, emp := range subordinates {
		// Count subordinates of this employee
		var subCount int
		var subs []struct{ ID string }
		h.DB.From("employees").Select("id").Eq("manager_id", emp.ID).Execute(&subs)
		subCount = len(subs)

		// Get task stats
		var tasks []struct {
			Status  string  `json:"status"`
			DueDate *string `json:"due_date"`
		}
		h.DB.From("tasks").Select("status, due_date").Eq("assignee_id", emp.ID).Execute(&tasks)

		openTasks := 0
		overdueTasks := 0
		for _, t := range tasks {
			if t.Status != "done" {
				openTasks++
				if t.DueDate != nil && *t.DueDate < today {
					overdueTasks++
				}
			}
		}

		// Get meetings count (from meetings table for this month)
		startOfMonth := time.Now().Format("2006-01") + "-01"
		var meetings []struct{ ID string }
		h.DB.From("meetings").Select("id").Eq("employee_id", emp.ID).Gte("date", startOfMonth).Execute(&meetings)

		result = append(result, fiber.Map{
			"id":            emp.ID,
			"name":          emp.Name,
			"position":      emp.Position,
			"photo_base64":  emp.PhotoBase64,
			"subordinates":  subCount,
			"open_tasks":    openTasks,
			"overdue_tasks": overdueTasks,
			"meetings":      len(meetings),
		})
	}

	return c.JSON(result)
}
