package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type RateLimiter struct {
	max    int
	window time.Duration
	cache  *cache.Cache
	mu     sync.Mutex
}

func New(window time.Duration, max int) *RateLimiter {
	return &RateLimiter{
		max:    max,
		window: window,
		cache:  cache.New(2*window, window),
	}
}

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowKey := ip + ":" + strconv.FormatInt(now.Unix()/int64(rl.window.Seconds()), 10)

	raw, found := rl.cache.Get(windowKey)
	current := 0
	if found {
		current = raw.(int)
	}

	if current >= rl.max {
		return false
	}

	rl.cache.Set(windowKey, current+1, rl.window)
	return true
}
