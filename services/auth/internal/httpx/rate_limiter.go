package httpx

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/response"
)

type RateLimiterConfig struct {
	Limit  int
	Window time.Duration
}

type rateLimitBucket struct {
	count      int
	windowEnds time.Time
}

type RateLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	buckets map[string]rateLimitBucket
}

func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	if cfg.Limit <= 0 {
		cfg.Limit = 20
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}

	return &RateLimiter{
		limit:   cfg.Limit,
		window:  cfg.Window,
		buckets: make(map[string]rateLimitBucket),
	}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		allowed, retryAfter := l.allow(key, time.Now())
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			response.Error(w, http.StatusTooManyRequests, "too many requests", map[string]any{
				"retry_after_seconds": int(retryAfter.Seconds()),
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket := l.buckets[key]
	if bucket.windowEnds.IsZero() || now.After(bucket.windowEnds) {
		l.buckets[key] = rateLimitBucket{
			count:      1,
			windowEnds: now.Add(l.window),
		}
		return true, 0
	}

	if bucket.count >= l.limit {
		return false, time.Until(bucket.windowEnds).Truncate(time.Second)
	}

	bucket.count++
	l.buckets[key] = bucket
	return true, 0
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
