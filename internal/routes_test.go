package routes

import (
	"net/http"
	"net/http/httptest"
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
