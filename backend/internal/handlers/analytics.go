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

	// Get employees
	var employees []models.Employee
	h.DB.From("employees").Select("*").Execute(&employees)

	// Get active projects
	var projects []models.Project
	h.DB.From("projects").Select("*").Eq("status", "active").Execute(&projects)

	// Get recent meetings
	var meetings []models.Meeting
	h.DB.From("meetings").Select("*, employees(name), meeting_categories(name)").Order("date", true).Limit(10).Execute(&meetings)

	// Get task stats
	var tasks []struct {
		Status  string  `json:"status"`
		DueDate *string `json:"due_date"`
	}
	h.DB.From("tasks").Select("status, due_date").Execute(&tasks)

	today := time.Now().Format("2006-01-02")
	totalTasks := len(tasks)
	doneTasks := 0
	inProgressTasks := 0
	overdueTasks := 0

	for _, t := range tasks {
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
		if m.Date >= startOfMonth {
			meetingsThisMonth++
		}
	}

	// Meetings by category
	meetingsByCategory := make(map[string]int)
	for _, m := range meetings {
		if m.Category != nil {
			meetingsByCategory[m.Category.Code]++
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
		"agreements_total":            0,
		"agreements_completed":        0,
		"agreements_overdue":          0,
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

	employeeID := c.Params("id")

	// Get meetings
	var meetings []models.Meeting
	h.DB.From("meetings").Select("id, date, mood_score, analysis").Eq("employee_id", employeeID).Order("date", false).Execute(&meetings)

	// Mood history
	var moodHistory []fiber.Map
	for _, m := range meetings {
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
		Status string `json:"status"`
	}
	h.DB.From("tasks").Select("status").Eq("assignee_id", employeeID).Execute(&tasks)

	totalTasks := len(tasks)
	doneTasks := 0
	inProgressTasks := 0
	for _, t := range tasks {
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
		if a.Status == "completed" {
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
		"total_meetings": len(meetings),
	})
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
			"category":            cat,
			"meetings_count":      len(catMeetings),
			"avg_mood":            avgMood,
			"mood_trend":          nil,
			"agreements_created":  0,
			"agreements_completed": 0,
			"common_topics":       []string{},
			"red_flags_count":     redFlagsCount,
			"specific_metrics":    fiber.Map{},
		})
	}

	return c.JSON(result)
}
