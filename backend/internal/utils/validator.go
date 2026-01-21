package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validator provides input validation methods
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

// Required validates that a string is not empty
func (v *Validator) Required(field, value string) bool {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
		return false
	}
	return true
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) bool {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, fmt.Sprintf("must be at least %d characters", min))
		return false
	}
	return true
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) bool {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("must be at most %d characters", max))
		return false
	}
	return true
}

// UUID validates UUID format
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (v *Validator) UUID(field, value string) bool {
	if value == "" {
		return true // Use Required for empty check
	}
	if !uuidRegex.MatchString(value) {
		v.AddError(field, "must be a valid UUID")
		return false
	}
	return true
}

// Email validates email format
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (v *Validator) Email(field, value string) bool {
	if value == "" {
		return true // Use Required for empty check
	}
	if !emailRegex.MatchString(value) {
		v.AddError(field, "must be a valid email address")
		return false
	}
	return true
}

// NoSQLInjection checks for potential SQL injection patterns
var sqlInjectionPatterns = regexp.MustCompile(`(?i)(--|;|'|"|\\|\/\*|\*\/|xp_|sp_|0x|union\s+select|insert\s+into|delete\s+from|drop\s+table|update\s+set)`)

func (v *Validator) NoSQLInjection(field, value string) bool {
	if sqlInjectionPatterns.MatchString(value) {
		v.AddError(field, "contains invalid characters")
		return false
	}
	return true
}

// NoXSS checks for potential XSS patterns
var xssPatterns = regexp.MustCompile(`(?i)(<script|javascript:|on\w+\s*=|<iframe|<object|<embed|<form|<input|<button)`)

func (v *Validator) NoXSS(field, value string) bool {
	if xssPatterns.MatchString(value) {
		v.AddError(field, "contains potentially unsafe content")
		return false
	}
	return true
}

// SafeString validates a string is safe (no SQL injection or XSS)
func (v *Validator) SafeString(field, value string) bool {
	sqlOk := v.NoSQLInjection(field, value)
	xssOk := v.NoXSS(field, value)
	return sqlOk && xssOk
}

// InList validates that a value is in a list of allowed values
func (v *Validator) InList(field, value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.AddError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
	return false
}

// Positive validates that an integer is positive
func (v *Validator) Positive(field string, value int) bool {
	if value <= 0 {
		v.AddError(field, "must be a positive number")
		return false
	}
	return true
}

// Range validates that an integer is within a range
func (v *Validator) Range(field string, value, min, max int) bool {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("must be between %d and %d", min, max))
		return false
	}
	return true
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// Trim whitespace
	s = strings.TrimSpace(s)
	return s
}

// SanitizeHTML escapes HTML special characters
func SanitizeHTML(s string) string {
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}
