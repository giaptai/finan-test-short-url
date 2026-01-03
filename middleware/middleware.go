package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           10800,
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// check if origin is allow
		allowOrigin := ""
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
		}
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if c.Request.Method == "OPTIONS" {
				c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowMethods))
				c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders))
				if config.MaxAge > 0 {
					c.Header("Access-Control-Max-Age", intToString(config.MaxAge))
				}
				c.AbortWithStatus(204)
				return
			}

			if len(config.ExposeHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders))
			}
		}
		c.Next()
	}
}

func joinStrings(s []string) string {
	if len(s) == 0 {
		return ""
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result += ", " + s[i]
	}
	return result
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int           // Maximum requests allowed
	Window   time.Duration // Time window for the limit
}

// client tracks request count for a single IP
type client struct {
	count    int
	lastSeen time.Time
}

// rateLimiter manages rate limiting state
type rateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	config  RateLimitConfig
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*client),
		config:  config,
	}
	// Cleanup old entries periodically
	go rl.cleanup()
	return rl
}

// cleanup removes expired entries
func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.config.Window)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > rl.config.Window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// isAllowed checks if the IP is allowed to make a request
func (rl *rateLimiter) isAllowed(ip string) (bool, int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		rl.clients[ip] = &client{count: 1, lastSeen: time.Now()}
		return true, rl.config.Requests - 1, rl.config.Window
	}

	// Reset if window has passed
	if time.Since(c.lastSeen) > rl.config.Window {
		c.count = 1
		c.lastSeen = time.Now()
		return true, rl.config.Requests - 1, rl.config.Window
	}

	// Check if limit exceeded
	if c.count >= rl.config.Requests {
		retryAfter := rl.config.Window - time.Since(c.lastSeen)
		return false, 0, retryAfter
	}

	c.count++
	remaining := rl.config.Requests - c.count
	return true, remaining, rl.config.Window - time.Since(c.lastSeen)
}

// DefaultRateLimitConfig returns default rate limit: 100 requests per minute
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
	}
}

// RateLimit returns a rate limiting middleware with default config
func RateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig())
}

// RateLimitWithConfig returns a rate limiting middleware with custom config
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		allowed, remaining, retryAfter := limiter.isAllowed(ip)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(retryAfter.Seconds()),
			})
			return
		}

		c.Next()
	}
}
