package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Clean up old entries periodically
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Filter requests within the window
	var valid []time.Time
	for _, t := range rl.requests[key] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// RateLimitByIP creates a rate limiting middleware using client IP
func RateLimitByIP(limit int, window time.Duration) fiber.Handler {
	limiter := NewRateLimiter(limit, window)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if !limiter.Allow(ip) {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Too many requests",
				"retry_after": int(window.Seconds()),
			})
		}
		return c.Next()
	}
}

// AuthRateLimiter is specifically for auth endpoints with stricter limits
func AuthRateLimiter() fiber.Handler {
	// 5 requests per minute per IP for auth endpoints
	return RateLimitByIP(5, time.Minute)
}

// APIRateLimiter is for general API endpoints
func APIRateLimiter() fiber.Handler {
	// 100 requests per minute per IP for general API
	return RateLimitByIP(100, time.Minute)
}
