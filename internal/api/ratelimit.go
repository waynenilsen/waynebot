package api

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
	rateLimit float64 // tokens per second
	burst     int
}

// RateLimiter provides per-IP rate limiting middleware.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor

	authRate  float64
	authBurst int

	defaultRate  float64
	defaultBurst int
}

// NewRateLimiter creates a rate limiter with separate limits for auth and other endpoints.
// authPerMin is the limit for /api/auth/* routes, otherPerMin for everything else.
func NewRateLimiter(authPerMin, otherPerMin int) *RateLimiter {
	rl := &RateLimiter{
		visitors:     make(map[string]*visitor),
		authRate:     float64(authPerMin) / 60.0,
		authBurst:    authPerMin,
		defaultRate:  float64(otherPerMin) / 60.0,
		defaultBurst: otherPerMin,
	}
	go rl.cleanup()
	return rl
}

// Middleware returns an HTTP middleware that enforces rate limits.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		isAuth := strings.HasPrefix(r.URL.Path, "/api/auth/")

		key := ip
		if isAuth {
			key += ":auth"
		}

		rl.mu.Lock()
		v, ok := rl.visitors[key]
		if !ok {
			rate := rl.defaultRate
			burst := rl.defaultBurst
			if isAuth {
				rate = rl.authRate
				burst = rl.authBurst
			}
			v = &visitor{
				tokens:    float64(burst),
				lastSeen:  time.Now(),
				rateLimit: rate,
				burst:     burst,
			}
			rl.visitors[key] = v
		}

		now := time.Now()
		elapsed := now.Sub(v.lastSeen).Seconds()
		v.lastSeen = now
		v.tokens += elapsed * v.rateLimit
		if v.tokens > float64(v.burst) {
			v.tokens = float64(v.burst)
		}

		if v.tokens < 1 {
			rl.mu.Unlock()
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		v.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// cleanup removes stale visitors every 5 minutes.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}
