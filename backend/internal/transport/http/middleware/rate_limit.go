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

type RateLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	buckets map[string]*rateLimitBucket
}

func NewRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &RateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*rateLimitBucket),
	}

	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()
		if !limiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}

func (r *RateLimiter) allow(key string) bool {
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
