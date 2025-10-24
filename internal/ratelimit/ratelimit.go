package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter for a specific endpoint
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitManager manages rate limiters for different IPs
type RateLimitManager struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(rateLimit rate.Limit, burst int) *RateLimitManager {
	manager := &RateLimitManager{
		limiters: make(map[string]*RateLimiter),
		rate:     rateLimit,
		burst:    burst,
	}

	// Start cleanup goroutine to remove old limiters
	go manager.cleanup()

	return manager
}

// GetLimiter gets or creates a rate limiter for an IP address
func (rlm *RateLimitManager) GetLimiter(ip string) *rate.Limiter {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	limiter, exists := rlm.limiters[ip]
	if !exists {
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(rlm.rate, rlm.burst),
			lastSeen: time.Now(),
		}
		rlm.limiters[ip] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter
}

// cleanup removes old rate limiters to prevent memory leaks
func (rlm *RateLimitManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rlm.mu.Lock()
		now := time.Now()
		for ip, limiter := range rlm.limiters {
			if now.Sub(limiter.lastSeen) > 10*time.Minute {
				delete(rlm.limiters, ip)
			}
		}
		rlm.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(manager *RateLimitManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP address
			ip := getClientIP(r)

			// Get rate limiter for this IP
			limiter := manager.GetLimiter(ip)

			// Check if request is allowed
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if len(xff) > 0 {
			return xff
		}
	}

	// Check X-Real-IP header (for nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if ip == "" {
		return "unknown"
	}

	return ip
}

// Predefined rate limiters for common use cases
var (
	// LoginRateLimit: 5 requests per minute (for login attempts)
	LoginRateLimit = NewRateLimitManager(rate.Every(time.Minute/5), 5)

	// RegistrationRateLimit: 3 requests per hour (for registration)
	RegistrationRateLimit = NewRateLimitManager(rate.Every(time.Hour/3), 3)

	// PasswordResetRateLimit: 3 requests per hour (for password reset)
	PasswordResetRateLimit = NewRateLimitManager(rate.Every(time.Hour/3), 3)

	// VerificationResendRateLimit: 5 requests per hour (for resending verification)
	VerificationResendRateLimit = NewRateLimitManager(rate.Every(time.Hour/5), 5)

	// GeneralAPIRateLimit: 100 requests per minute (for general API usage)
	GeneralAPIRateLimit = NewRateLimitManager(rate.Every(time.Minute/100), 100)
)

// Middleware functions for common endpoints
func LoginRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(LoginRateLimit)
}

func RegistrationRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(RegistrationRateLimit)
}

func PasswordResetRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(PasswordResetRateLimit)
}

func VerificationResendRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(VerificationResendRateLimit)
}

func GeneralAPIRateLimitMiddleware() func(http.Handler) http.Handler {
	return RateLimitMiddleware(GeneralAPIRateLimit)
}
