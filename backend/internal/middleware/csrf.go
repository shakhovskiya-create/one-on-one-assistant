package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	csrfTokenLength = 32
	csrfCookieName  = "csrf_token"
	csrfHeaderName  = "X-CSRF-Token"
	csrfTokenExpiry = 24 * time.Hour
)

// CSRFToken represents a CSRF token with expiry
type CSRFToken struct {
	Token   string
	Created time.Time
}

// CSRFStore stores CSRF tokens
type CSRFStore struct {
	tokens map[string]CSRFToken
	mu     sync.RWMutex
}

// NewCSRFStore creates a new CSRF token store
func NewCSRFStore() *CSRFStore {
	store := &CSRFStore{
		tokens: make(map[string]CSRFToken),
	}
	// Clean up expired tokens periodically
	go store.cleanup()
	return store
}

func (s *CSRFStore) cleanup() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, token := range s.tokens {
			if now.Sub(token.Created) > csrfTokenExpiry {
				delete(s.tokens, key)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateToken generates a new CSRF token for a session
func (s *CSRFStore) GenerateToken(sessionID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokenBytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		// Fallback - should never happen
		return ""
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	s.tokens[sessionID] = CSRFToken{
		Token:   token,
		Created: time.Now(),
	}
	return token
}

// ValidateToken validates a CSRF token
func (s *CSRFStore) ValidateToken(sessionID, token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stored, exists := s.tokens[sessionID]
	if !exists {
		return false
	}

	// Check expiry
	if time.Since(stored.Created) > csrfTokenExpiry {
		return false
	}

	// Constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(stored.Token), []byte(token)) == 1
}

var csrfStore = NewCSRFStore()

// CSRFProtection provides CSRF protection middleware
func CSRFProtection() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip CSRF for safe methods
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// Get session ID from JWT or cookie
		sessionID := ""
		if userID := c.Locals("user_id"); userID != nil {
			sessionID = userID.(string)
		}
		if sessionID == "" {
			// No session, skip CSRF (unauthenticated requests)
			return c.Next()
		}

		// Validate CSRF token from header
		token := c.Get(csrfHeaderName)
		if token == "" {
			// Also check form field for traditional forms
			token = c.FormValue("_csrf")
		}

		if token == "" {
			return c.Status(403).JSON(fiber.Map{
				"error": "CSRF token missing",
			})
		}

		if !csrfStore.ValidateToken(sessionID, token) {
			return c.Status(403).JSON(fiber.Map{
				"error": "Invalid CSRF token",
			})
		}

		return c.Next()
	}
}

// CSRFTokenHandler returns a new CSRF token for the current session
func CSRFTokenHandler(c *fiber.Ctx) error {
	sessionID := ""
	if userID := c.Locals("user_id"); userID != nil {
		sessionID = userID.(string)
	}
	if sessionID == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	token := csrfStore.GenerateToken(sessionID)
	return c.JSON(fiber.Map{
		"csrf_token": token,
	})
}
