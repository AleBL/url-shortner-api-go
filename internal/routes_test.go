package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router)
	return router
}

func request(router *gin.Engine, method, path, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if method == http.MethodOptions {
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func shortenFrom(router *gin.Engine, remoteAddr string, headers map[string]string) int {
	req := httptest.NewRequest(http.MethodPost, "/shorten", nil)
	req.RemoteAddr = remoteAddr
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

func TestShortenRateLimitUsesClientIPHeader(t *testing.T) {
	t.Setenv("CLIENT_IP_HEADER", "True-Client-IP")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "1")
	router := newTestRouter()
	proxy := "10.0.0.1:1000"

	if code := shortenFrom(router, proxy, map[string]string{"True-Client-IP": "198.51.100.1"}); code == http.StatusTooManyRequests {
		t.Fatal("first request from client 1 was rate limited")
	}
	if code := shortenFrom(router, proxy, map[string]string{"True-Client-IP": "198.51.100.1"}); code != http.StatusTooManyRequests {
		t.Fatalf("second request from client 1: status = %d, want %d", code, http.StatusTooManyRequests)
	}
	if code := shortenFrom(router, proxy, map[string]string{"True-Client-IP": "198.51.100.2"}); code == http.StatusTooManyRequests {
		t.Fatal("client 2 was limited by client 1's requests")
	}
}

func TestShortenRateLimitIgnoresForgedForwardedFor(t *testing.T) {
	t.Setenv("CLIENT_IP_HEADER", "")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "1")
	router := newTestRouter()

	shortenFrom(router, "198.51.100.1:1000", map[string]string{"X-Forwarded-For": "203.0.113.1"})
	code := shortenFrom(router, "198.51.100.1:1000", map[string]string{"X-Forwarded-For": "203.0.113.2"})
	if code != http.StatusTooManyRequests {
		t.Fatalf("forged X-Forwarded-For bypassed the limit: status = %d", code)
	}
}

func TestDebugClientIP(t *testing.T) {
	t.Setenv("DEBUG_CLIENT_IP", "")
	if rec := request(newTestRouter(), http.MethodGet, "/debug/client-ip", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("disabled: status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	t.Setenv("DEBUG_CLIENT_IP", "true")
	rec := request(newTestRouter(), http.MethodGet, "/debug/client-ip", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"client_ip"`) {
		t.Fatalf("enabled: status = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com/, http://localhost:5173")
	router := newTestRouter()

	for _, origin := range []string{"https://app.example.com", "http://localhost:5173"} {
		rec := request(router, http.MethodOptions, "/shorten", origin)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("preflight from %s: status = %d, want %d", origin, rec.Code, http.StatusNoContent)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("preflight from %s: Access-Control-Allow-Origin = %q", origin, got)
		}
	}

	rec := request(router, http.MethodGet, "/health", "https://app.example.com")
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("GET: Access-Control-Allow-Origin = %q", got)
	}
}

func TestCORSExposesRetryAfterOnRateLimit(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "1")
	router := newTestRouter()

	request(router, http.MethodPost, "/shorten", "https://app.example.com")
	rec := request(router, http.MethodPost, "/shorten", "https://app.example.com")

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Expose-Headers"); !strings.Contains(strings.ToLower(got), "retry-after") {
		t.Fatalf("Access-Control-Expose-Headers = %q, want it to include Retry-After", got)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header is missing")
	}
}

func TestCORSRejectsOtherOrigins(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
	router := newTestRouter()

	rec := request(router, http.MethodOptions, "/shorten", "https://evil.example.com")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCORSDisabledWithoutOrigins(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")
	router := newTestRouter()

	rec := request(router, http.MethodGet, "/health", "https://app.example.com")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}
