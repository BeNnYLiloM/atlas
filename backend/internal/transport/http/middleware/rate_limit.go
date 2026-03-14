package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// slidingWindowBucket хранит временные метки запросов в пределах окна.
type slidingWindowBucket struct {
	timestamps []time.Time
}

type rateLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	buckets map[string]*slidingWindowBucket
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*slidingWindowBucket),
	}
	go rl.cleanupLoop()
	return rl
}

func (r *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		r.cleanup()
	}
}

func (r *rateLimiter) cleanup() {
	now := time.Now().UTC()
	cutoff := now.Add(-r.window)
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, bucket := range r.buckets {
		// Оставляем только свежие метки
		fresh := bucket.timestamps[:0]
		for _, ts := range bucket.timestamps {
			if ts.After(cutoff) {
				fresh = append(fresh, ts)
			}
		}
		if len(fresh) == 0 {
			delete(r.buckets, key)
		} else {
			bucket.timestamps = fresh
		}
	}
}

// allow реализует sliding window: считает запросы за последние window секунд.
func (r *rateLimiter) allow(key string) bool {
	now := time.Now().UTC()
	cutoff := now.Add(-r.window)

	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, exists := r.buckets[key]
	if !exists {
		r.buckets[key] = &slidingWindowBucket{timestamps: []time.Time{now}}
		return true
	}

	// Отсекаем устаревшие метки
	fresh := bucket.timestamps[:0]
	for _, ts := range bucket.timestamps {
		if ts.After(cutoff) {
			fresh = append(fresh, ts)
		}
	}
	bucket.timestamps = fresh

	if len(bucket.timestamps) >= r.limit {
		return false
	}

	bucket.timestamps = append(bucket.timestamps, now)
	return true
}

// getUserIDSafe возвращает userID из контекста без паники, или пустую строку.
func getUserIDSafe(c *gin.Context) string {
	v, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	id, _ := v.(string)
	return id
}

// NewRateLimiter возвращает middleware с лимитом по userID (аутентифицированный запрос).
// Если userID недоступен (нет токена) — фолбэк на IP.
func NewRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		// Используем userID как ключ — точный лимит per user, не per IP (NAT-safe)
		key := getUserIDSafe(c)
		if key == "" {
			key = c.ClientIP()
		}
		key += ":" + c.FullPath()

		if !rl.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}
