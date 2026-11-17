package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	idleClientTTL   = 3 * time.Minute
	cleanupInterval = time.Minute
)

// RateLimit allows each client IP perMinute requests per minute, with bursts of
// the same size. State lives in memory, which fits a single instance and resets
// on restart. perMinute <= 0 disables the limit.
func RateLimit(perMinute int) gin.HandlerFunc {
	if perMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	clients := &clientLimiters{
		limit:    rate.Limit(float64(perMinute) / 60),
		burst:    perMinute,
		limiters: make(map[string]*clientLimiter),
	}
	retryAfter := strconv.Itoa(int(math.Ceil(60 / float64(perMinute))))

	return func(c *gin.Context) {
		if !clients.allow(c.ClientIP(), time.Now()) {
			c.Header("Retry-After", retryAfter)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type clientLimiters struct {
	mu          sync.Mutex
	limit       rate.Limit
	burst       int
	limiters    map[string]*clientLimiter
	lastCleanup time.Time
}

func (l *clientLimiters) allow(ip string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Forget idle clients so the map does not grow forever; their buckets would be full again anyway.
	if now.Sub(l.lastCleanup) > cleanupInterval {
		for key, client := range l.limiters {
			if now.Sub(client.lastSeen) > idleClientTTL {
				delete(l.limiters, key)
			}
		}
		l.lastCleanup = now
	}

	client, ok := l.limiters[ip]
	if !ok {
		client = &clientLimiter{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.limiters[ip] = client
	}
	client.lastSeen = now
	return client.limiter.AllowN(now, 1)
}
