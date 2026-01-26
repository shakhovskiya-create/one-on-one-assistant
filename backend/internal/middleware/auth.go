package middleware

import (
	"strings"
	"time"

	"github.com/ekf/one-on-one-backend/pkg/auth"
	"github.com/gofiber/fiber/v2"
)

// Cookie names
const (
	AuthCookieName    = "ekf_auth_token"
	AuthCookieMaxAge  = 24 * 60 * 60 // 24 hours in seconds
)

// SetAuthCookie sets the HttpOnly authentication cookie
func SetAuthCookie(c *fiber.Ctx, token string, secure bool) {
	cookie := fiber.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,                    // Prevents JavaScript access
		Secure:   secure,                  // Only send over HTTPS in production
		SameSite: "Lax",                   // CSRF protection
		Path:     "/",
	}
	c.Cookie(&cookie)
}

// ClearAuthCookie removes the authentication cookie
func ClearAuthCookie(c *fiber.Ctx) {
	cookie := fiber.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Expired
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	}
	c.Cookie(&cookie)
}

// JWTAuth creates a middleware for JWT authentication
// Supports both Authorization header (for API clients) and HttpOnly cookie (for browser)
func JWTAuth(jwtManager *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		// 1. Try Authorization header first (for API clients, mobile apps)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.Status(401).JSON(fiber.Map{"error": "Invalid authorization header format"})
			}
		}

		// 2. Fall back to HttpOnly cookie (for browser)
		if tokenString == "" {
			tokenString = c.Cookies(AuthCookieName)
		}

		// 3. No token found
		if tokenString == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Authentication required"})
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			// Clear invalid cookie
			ClearAuthCookie(c)
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("name", claims.Name)
		c.Locals("department", claims.Department)

		return c.Next()
	}
}

// OptionalJWTAuth validates JWT if present but doesn't require it
// Supports both Authorization header and HttpOnly cookie
func OptionalJWTAuth(jwtManager *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		// 1. Try Authorization header first
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				tokenString = "" // Invalid format, ignore
			}
		}

		// 2. Fall back to HttpOnly cookie
		if tokenString == "" {
			tokenString = c.Cookies(AuthCookieName)
		}

		// 3. Validate if token found
		if tokenString != "" {
			claims, err := jwtManager.ValidateToken(tokenString)
			if err == nil {
				c.Locals("user_id", claims.UserID)
				c.Locals("email", claims.Email)
				c.Locals("name", claims.Name)
				c.Locals("department", claims.Department)
			}
		}

		return c.Next()
	}
}
