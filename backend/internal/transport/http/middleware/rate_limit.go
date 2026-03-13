package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitBucket struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	buckets map[string]*rateLimitBucket
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*rateLimitBucket),
	}
	go rl.cleanupLoop()
	return rl
}

// cleanupLoop периодически удаляет истёкшие buckets чтобы не было утечки памяти.
func (r *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		r.cleanup()
	}
}

func (r *rateLimiter) cleanup() {
	now := time.Now().UTC()
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, bucket := range r.buckets {
		if now.After(bucket.resetAt) {
			delete(r.buckets, key)
		}
	}
}

func (r *rateLimiter) allow(key string) bool {
	now := time.Now().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, exists := r.buckets[key]
	if !exists || now.After(bucket.resetAt) {
		r.buckets[key] = &rateLimitBucket{
			count:   1,
			resetAt: now.Add(r.window),
		}
		return true
	}

	if bucket.count >= r.limit {
		return false
	}

	bucket.count++
	return true
}

func NewRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()
		if !rl.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}
