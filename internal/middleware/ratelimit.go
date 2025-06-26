package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	visitors    map[string]*rate.Limiter
	mu          sync.RWMutex
	r           rate.Limit
	b           int
	cleanupInt  time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
	cleanupOnce sync.Once
}

func NewRateLimiter(r rate.Limit, b int, cleanupInterval time.Duration) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		visitors:   make(map[string]*rate.Limiter),
		r:          r,
		b:          b,
		cleanupInt: cleanupInterval,
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Start cleanup goroutine only once
	go rl.cleanupVisitors()
	
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.visitors[ip] = limiter
	}

	return limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()
	
	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			for ip, limiter := range rl.visitors {
				if limiter.Allow() {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Close stops the cleanup goroutine
func (rl *RateLimiter) Close() {
	rl.cancel()
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the real client IP from the request
func getIP(r *http.Request) string {
	// Check X-Real-IP first (nginx proxy)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return cleanIP(ip)
	}
	
	// Check X-Forwarded-For (can contain multiple IPs)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the list
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return cleanIP(forwarded[:idx])
		}
		return cleanIP(forwarded)
	}
	
	// Fall back to RemoteAddr
	return cleanIP(r.RemoteAddr)
}

// cleanIP removes port from IP address if present
func cleanIP(ip string) string {
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		// IPv6 addresses have multiple colons, only strip port if it's clearly a port
		if !strings.Contains(ip[:idx], ":") {
			return strings.TrimSpace(ip[:idx])
		}
	}
	return strings.TrimSpace(ip)
}