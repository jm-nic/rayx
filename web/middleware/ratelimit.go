package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipEntry holds a rate limiter and the last time a request was seen from an IP.
type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages per-IP token-bucket rate limiters.
type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	r       rate.Limit
	burst   int
	done    chan struct{}
}

// NewIPRateLimiter creates a new per-IP rate limiter.
// r is the sustained request rate (e.g. rate.Every(2*time.Second) = 0.5 req/s = 30 req/min).
// burst is the maximum number of requests allowed in a short burst.
// Call Stop() to release the background cleanup goroutine when the limiter is no longer needed.
func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		entries: make(map[string]*ipEntry),
		r:       r,
		burst:   burst,
		done:    make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Stop terminates the background cleanup goroutine. It should be called when
// the limiter will no longer be used (e.g. during graceful server shutdown).
func (rl *IPRateLimiter) Stop() {
	close(rl.done)
}

// getLimiter returns the rate.Limiter for the given IP, creating one if it does not exist.
func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if e, ok := rl.entries[ip]; ok {
		e.lastSeen = time.Now()
		return e.limiter
	}
	e := &ipEntry{
		limiter:  rate.NewLimiter(rl.r, rl.burst),
		lastSeen: time.Now(),
	}
	rl.entries[ip] = e
	return e.limiter
}

// cleanupLoop periodically removes entries for IPs that have not been seen recently,
// preventing unbounded memory growth. It exits when Stop() is called.
func (rl *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, e := range rl.entries {
				if time.Since(e.lastSeen) > 10*time.Minute {
					delete(rl.entries, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// RateLimit returns a Gin middleware that enforces the per-IP rate limit.
// Requests that exceed the limit receive HTTP 429 Too Many Requests.
func RateLimit(rl *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"msg":     "too many requests",
			})
			return
		}
		c.Next()
	}
}
