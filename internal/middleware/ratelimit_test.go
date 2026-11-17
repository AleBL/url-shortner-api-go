package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newRouter(perMinute int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", RateLimit(perMinute), func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

func get(router *gin.Engine, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRateLimitBlocksAfterBurst(t *testing.T) {
	router := newRouter(2)

	for i := 0; i < 2; i++ {
		if rec := get(router, "198.51.100.1:1000"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i+1, rec.Code, http.StatusOK)
		}
	}

	rec := get(router, "198.51.100.1:1000")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got != "30" {
		t.Fatalf("Retry-After = %q, want %q", got, "30")
	}

	if rec := get(router, "198.51.100.2:1000"); rec.Code != http.StatusOK {
		t.Fatalf("other client: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRateLimitDisabled(t *testing.T) {
	router := newRouter(0)

	for i := 0; i < 50; i++ {
		if rec := get(router, "198.51.100.1:1000"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i+1, rec.Code, http.StatusOK)
		}
	}
}

func TestClientLimitersForgetIdleClients(t *testing.T) {
	clients := &clientLimiters{limit: 1, burst: 1, limiters: make(map[string]*clientLimiter)}
	start := time.Now()

	clients.allow("198.51.100.1", start)
	clients.allow("198.51.100.2", start.Add(idleClientTTL+time.Second))

	if _, ok := clients.limiters["198.51.100.1"]; ok {
		t.Fatal("idle client was not removed")
	}
	if len(clients.limiters) != 1 {
		t.Fatalf("limiters = %d, want 1", len(clients.limiters))
	}
}
