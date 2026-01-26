package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ekf/one-on-one-backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ListImprovementRequests returns improvement requests with filters
func (h *Handler) ListImprovementRequests(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	query := h.DB.From("improvement_requests").Select("*, initiator:employees!improvement_requests_initiator_id_fkey(id, name, photo_base64), sponsor:employees!improvement_requests_sponsor_id_fkey(id, name, photo_base64), type:improvement_request_types(id, name, icon, color)")

	// Filters
	if initiatorID := c.Query("initiator_id"); initiatorID != "" {
		query = query.Eq("initiator_id", initiatorID)
	}
	if sponsorID := c.Query("sponsor_id"); sponsorID != "" {
		query = query.Eq("sponsor_id", sponsorID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Eq("status", status)
	}
	if departmentID := c.Query("department_id"); departmentID != "" {
		query = query.Eq("department_id", departmentID)
	}
	if typeID := c.Query("type_id"); typeID != "" {
		query = query.Eq("type_id", typeID)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Eq("priority", priority)
	}

	var requests []models.ImprovementRequest
	err := query.Order("created_at", true).Execute(&requests)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(requests)
}

// GetImprovementRequest returns a single improvement request with details
func (h *Handler) GetImprovementRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var request models.ImprovementRequest
	err := h.DB.From("improvement_requests").
		Select("*, initiator:employees!improvement_requests_initiator_id_fkey(id, name, email, position, department, photo_base64), sponsor:employees!improvement_requests_sponsor_id_fkey(id, name, email, position, department, photo_base64), type:improvement_request_types(id, name, description, icon, color), project:projects(id, name, status)").
		Eq("id", id).Single().Execute(&request)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Improvement request not found"})
	}

	// Get comments
	var comments []models.ImprovementRequestComment
	h.DB.From("improvement_request_comments").Select("*, author:employees(id, name, photo_base64)").Eq("request_id", id).Order("created_at", false).Execute(&comments)
	request.Comments = comments

	// Get approvals
	var approvals []models.ImprovementRequestApproval
	h.DB.From("improvement_request_approvals").Select("*, approver:employees(id, name, photo_base64)").Eq("request_id", id).Order("created_at", false).Execute(&approvals)
	request.Approvals = approvals

	// Get activity
	var activity []models.ImprovementRequestActivity
	h.DB.From("improvement_request_activity").Select("*, actor:employees(id, name)").Eq("request_id", id).Order("created_at", true).Limit(50).Execute(&activity)
	request.Activity = activity

	return c.JSON(request)
}

// CreateImprovementRequest creates a new improvement request
func (h *Handler) CreateImprovementRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var input struct {
		Title           string   `json:"title"`
		Description     *string  `json:"description"`
		BusinessValue   *string  `json:"business_value"`
		ExpectedEffect  *string  `json:"expected_effect"`
		InitiatorID     string   `json:"initiator_id"`
		DepartmentID    *string  `json:"department_id"`
		SponsorID       *string  `json:"sponsor_id"`
		EstimatedBudget *float64 `json:"estimated_budget"`
		EstimatedStart  *string  `json:"estimated_start"`
		EstimatedEnd    *string  `json:"estimated_end"`
		TypeID          *string  `json:"type_id"`
		Priority        string   `json:"priority"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Title == "" || input.InitiatorID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title and initiator_id required"})
	}

	// Default values
	if input.Priority == "" {
		input.Priority = "medium"
	}

	// Generate request number
	var lastRequest []struct {
		Number string `json:"number"`
	}
	h.DB.From("improvement_requests").Select("number").Ilike("number", "IR-2026-%").Order("created_at", true).Limit(1).Execute(&lastRequest)

	nextNum := 1
	if len(lastRequest) > 0 {
		var num int
		fmt.Sscanf(lastRequest[0].Number, "IR-2026-%d", &num)
		nextNum = num + 1
	}

	requestNumber := fmt.Sprintf("IR-2026-%04d", nextNum)

	data := map[string]interface{}{
		"number":           requestNumber,
		"title":            input.Title,
		"description":      input.Description,
		"business_value":   input.BusinessValue,
		"expected_effect":  input.ExpectedEffect,
		"initiator_id":     input.InitiatorID,
		"department_id":    input.DepartmentID,
		"sponsor_id":       input.SponsorID,
		"estimated_budget": input.EstimatedBudget,
		"estimated_start":  input.EstimatedStart,
		"estimated_end":    input.EstimatedEnd,
		"type_id":          input.TypeID,
		"priority":         input.Priority,
		"status":           "draft",
	}

	result, err := h.DB.Insert("improvement_requests", data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var created []models.ImprovementRequest
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(201).JSON(fiber.Map{"status": "created", "number": requestNumber})
	}

	// Log activity
	h.DB.Insert("improvement_request_activity", map[string]interface{}{
		"request_id": created[0].ID,
		"actor_id":   input.InitiatorID,
		"action":     "created",
	})

	return c.Status(201).JSON(created[0])
}

// UpdateImprovementRequest updates an improvement request
func (h *Handler) UpdateImprovementRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		Title             *string  `json:"title"`
		Description       *string  `json:"description"`
		BusinessValue     *string  `json:"business_value"`
		ExpectedEffect    *string  `json:"expected_effect"`
		SponsorID         *string  `json:"sponsor_id"`
		EstimatedBudget   *float64 `json:"estimated_budget"`
		ApprovedBudget    *float64 `json:"approved_budget"`
		EstimatedStart    *string  `json:"estimated_start"`
		EstimatedEnd      *string  `json:"estimated_end"`
		TypeID            *string  `json:"type_id"`
		Priority          *string  `json:"priority"`
		CommitteeDate     *string  `json:"committee_date"`
		CommitteeDecision *string  `json:"committee_decision"`
		ActorID           *string  `json:"actor_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	data := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if input.Title != nil {
		data["title"] = *input.Title
	}
	if input.Description != nil {
		data["description"] = *input.Description
	}
	if input.BusinessValue != nil {
		data["business_value"] = *input.BusinessValue
	}
	if input.ExpectedEffect != nil {
		data["expected_effect"] = *input.ExpectedEffect
	}
	if input.SponsorID != nil {
		data["sponsor_id"] = *input.SponsorID
	}
	if input.EstimatedBudget != nil {
		data["estimated_budget"] = *input.EstimatedBudget
	}
	if input.ApprovedBudget != nil {
		data["approved_budget"] = *input.ApprovedBudget
	}
	if input.EstimatedStart != nil {
		data["estimated_start"] = *input.EstimatedStart
	}
	if input.EstimatedEnd != nil {
		data["estimated_end"] = *input.EstimatedEnd
	}
	if input.TypeID != nil {
		data["type_id"] = *input.TypeID
	}
	if input.Priority != nil {
		data["priority"] = *input.Priority
	}
	if input.CommitteeDate != nil {
		data["committee_date"] = *input.CommitteeDate
	}
	if input.CommitteeDecision != nil {
		data["committee_decision"] = *input.CommitteeDecision
	}

	result, err := h.DB.Update("improvement_requests", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var updated []models.ImprovementRequest
	json.Unmarshal(result, &updated)
	if len(updated) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	return c.JSON(updated[0])
}

// SubmitImprovementRequest submits a draft request for review
func (h *Handler) SubmitImprovementRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		ActorID string `json:"actor_id"`
	}
	c.BodyParser(&input)

	// Get current request
	var request models.ImprovementRequest
	err := h.DB.From("improvement_requests").Eq("id", id).Single().Execute(&request)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	if request.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Only draft requests can be submitted"})
	}

	// Update status
	data := map[string]interface{}{
		"status":       "submitted",
		"submitted_at": time.Now(),
		"updated_at":   time.Now(),
	}

	result, err := h.DB.Update("improvement_requests", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	h.DB.Insert("improvement_request_activity", map[string]interface{}{
		"request_id": id,
		"actor_id":   input.ActorID,
		"action":     "submitted",
		"old_value":  "draft",
		"new_value":  "submitted",
	})

	var updated []models.ImprovementRequest
	json.Unmarshal(result, &updated)
	if len(updated) == 0 {
		return c.JSON(fiber.Map{"status": "submitted"})
	}

	return c.JSON(updated[0])
}

// ApproveImprovementRequest approves a request at the current stage
func (h *Handler) ApproveImprovementRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		ApproverID     string  `json:"approver_id"`
		Comment        *string `json:"comment"`
		ApprovedBudget *float64 `json:"approved_budget"` // For budgeting stage
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.ApproverID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "approver_id required"})
	}

	// Get current request
	var request models.ImprovementRequest
	err := h.DB.From("improvement_requests").Eq("id", id).Single().Execute(&request)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	// Determine next status based on current status
	statusTransitions := map[string]string{
		"submitted":        "screening",
		"screening":        "evaluation",
		"evaluation":       "manager_approval",
		"manager_approval": "committee_review",
		"committee_review": "budgeting",
		"budgeting":        "project_created",
	}

	nextStatus, ok := statusTransitions[request.Status]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "Request cannot be approved in current status: " + request.Status})
	}

	// Record approval
	h.DB.Insert("improvement_request_approvals", map[string]interface{}{
		"request_id":  id,
		"approver_id": input.ApproverID,
		"stage":       request.Status,
		"decision":    "approved",
		"comment":     input.Comment,
	})

	// Update status
	data := map[string]interface{}{
		"status":     nextStatus,
		"updated_at": time.Now(),
	}

	// If budgeting stage, update approved budget
	if request.Status == "budgeting" && input.ApprovedBudget != nil {
		data["approved_budget"] = *input.ApprovedBudget
		data["approved_at"] = time.Now()
	}

	result, err := h.DB.Update("improvement_requests", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	h.DB.Insert("improvement_request_activity", map[string]interface{}{
		"request_id": id,
		"actor_id":   input.ApproverID,
		"action":     "approved",
		"old_value":  request.Status,
		"new_value":  nextStatus,
	})

	var updated []models.ImprovementRequest
	json.Unmarshal(result, &updated)
	if len(updated) == 0 {
		return c.JSON(fiber.Map{"status": nextStatus})
	}

	return c.JSON(updated[0])
}

// RejectImprovementRequest rejects a request
func (h *Handler) RejectImprovementRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		RejectorID string `json:"rejector_id"`
		Reason     string `json:"reason"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.RejectorID == "" || input.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "rejector_id and reason required"})
	}

	// Get current request
	var request models.ImprovementRequest
	err := h.DB.From("improvement_requests").Eq("id", id).Single().Execute(&request)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	// Record rejection in approvals
	h.DB.Insert("improvement_request_approvals", map[string]interface{}{
		"request_id":  id,
		"approver_id": input.RejectorID,
		"stage":       request.Status,
		"decision":    "rejected",
		"comment":     input.Reason,
	})

	// Update status
	now := time.Now()
	data := map[string]interface{}{
		"status":           "rejected",
		"rejection_reason": input.Reason,
		"rejected_by":      input.RejectorID,
		"rejected_at":      now,
		"updated_at":       now,
	}

	result, err := h.DB.Update("improvement_requests", "id", id, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	h.DB.Insert("improvement_request_activity", map[string]interface{}{
		"request_id": id,
		"actor_id":   input.RejectorID,
		"action":     "rejected",
		"old_value":  request.Status,
		"new_value":  "rejected",
	})

	var updated []models.ImprovementRequest
	json.Unmarshal(result, &updated)
	if len(updated) == 0 {
		return c.JSON(fiber.Map{"status": "rejected"})
	}

	return c.JSON(updated[0])
}

// CreateProjectFromRequest creates a project from an approved improvement request
func (h *Handler) CreateProjectFromRequest(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	id := c.Params("id")

	var input struct {
		ActorID string `json:"actor_id"`
	}
	c.BodyParser(&input)

	// Get request
	var request models.ImprovementRequest
	err := h.DB.From("improvement_requests").Eq("id", id).Single().Execute(&request)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	if request.Status != "project_created" && request.Status != "budgeting" {
		return c.Status(400).JSON(fiber.Map{"error": "Request must be approved before creating project"})
	}

	if request.ProjectID != nil && *request.ProjectID != "" {
		return c.Status(400).JSON(fiber.Map{"error": "Project already created for this request"})
	}

	// Create project
	projectData := map[string]interface{}{
		"name":        request.Title,
		"description": request.Description,
		"status":      "planning",
		"start_date":  request.EstimatedStart,
		"end_date":    request.EstimatedEnd,
	}

	projectResult, err := h.DB.Insert("projects", projectData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create project: " + err.Error()})
	}

	var createdProjects []models.Project
	json.Unmarshal(projectResult, &createdProjects)
	if len(createdProjects) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create project"})
	}

	projectID := createdProjects[0].ID

	// Update request with project link
	updateData := map[string]interface{}{
		"project_id": projectID,
		"status":     "in_progress",
		"updated_at": time.Now(),
	}

	result, err := h.DB.Update("improvement_requests", "id", id, updateData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	h.DB.Insert("improvement_request_activity", map[string]interface{}{
		"request_id": id,
		"actor_id":   input.ActorID,
		"action":     "project_created",
		"new_value":  projectID,
	})

	var updated []models.ImprovementRequest
	json.Unmarshal(result, &updated)
	if len(updated) == 0 {
		return c.JSON(fiber.Map{"status": "project_created", "project_id": projectID})
	}

	updated[0].Project = &createdProjects[0]
	return c.JSON(updated[0])
}

// AddImprovementRequestComment adds a comment to a request
func (h *Handler) AddImprovementRequestComment(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	requestID := c.Params("id")
	var input struct {
		AuthorID   string `json:"author_id"`
		Content    string `json:"content"`
		IsInternal bool   `json:"is_internal"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.AuthorID == "" || input.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Author ID and content required"})
	}

	result, err := h.DB.Insert("improvement_request_comments", map[string]interface{}{
		"request_id":  requestID,
		"author_id":   input.AuthorID,
		"content":     input.Content,
		"is_internal": input.IsInternal,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	action := "comment_added"
	if input.IsInternal {
		action = "internal_note_added"
	}
	h.DB.Insert("improvement_request_activity", map[string]interface{}{
		"request_id": requestID,
		"actor_id":   input.AuthorID,
		"action":     action,
	})

	var created []models.ImprovementRequestComment
	json.Unmarshal(result, &created)
	if len(created) == 0 {
		return c.Status(201).JSON(fiber.Map{"status": "created"})
	}

	// Fetch with author info
	var comment models.ImprovementRequestComment
	err = h.DB.From("improvement_request_comments").
		Select("*, author:employees(id, name, photo_base64)").
		Eq("id", created[0].ID).
		Single().
		Execute(&comment)
	if err != nil {
		return c.Status(201).JSON(created[0])
	}

	return c.Status(201).JSON(comment)
}

// ListImprovementRequestTypes returns all types
func (h *Handler) ListImprovementRequestTypes(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	var types []models.ImprovementRequestType
	err := h.DB.From("improvement_request_types").Select("*").Order("name", false).Execute(&types)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(types)
}

// GetMyImprovementRequests returns requests for current user
func (h *Handler) GetMyImprovementRequests(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user_id required"})
	}

	var requests []models.ImprovementRequest
	err := h.DB.From("improvement_requests").
		Select("*, type:improvement_request_types(id, name, icon, color)").
		Eq("initiator_id", userID).
		Order("created_at", true).
		Limit(10).
		Execute(&requests)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(requests)
}

// GetImprovementRequestStats returns statistics
func (h *Handler) GetImprovementRequestStats(c *fiber.Ctx) error {
	if h.DB == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not configured"})
	}

	// Get counts by status
	var requests []models.ImprovementRequest
	h.DB.From("improvement_requests").Select("id, status").Execute(&requests)

	statusCounts := make(map[string]int)
	for _, r := range requests {
		statusCounts[r.Status]++
	}

	draft := statusCounts["draft"]
	pending := statusCounts["submitted"] + statusCounts["screening"] + statusCounts["evaluation"] + statusCounts["manager_approval"] + statusCounts["committee_review"] + statusCounts["budgeting"]
	approved := statusCounts["project_created"] + statusCounts["in_progress"] + statusCounts["completed"]
	rejected := statusCounts["rejected"]

	return c.JSON(fiber.Map{
		"total":    len(requests),
		"draft":    draft,
		"pending":  pending,
		"approved": approved,
		"rejected": rejected,
		"by_status": statusCounts,
	})
}
