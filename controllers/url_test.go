package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	routes "url-shortener/internal"
	"url-shortener/internal/database"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "url-shortener-controllers")
	if err != nil {
		panic(err)
	}
	os.Setenv("DB_PATH", filepath.Join(dir, "test.db"))
	gin.SetMode(gin.TestMode)

	code := m.Run()

	database.Close()
	os.RemoveAll(dir)
	os.Exit(code)
}

func newRouter() *gin.Engine {
	router := gin.New()
	routes.SetupRoutes(router)
	return router
}

func doRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func shorten(t *testing.T, router *gin.Engine, url string) string {
	t.Helper()
	rec := doRequest(router, http.MethodPost, "/shorten", `{"url":"`+url+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /shorten status = %d, body = %s", rec.Code, rec.Body)
	}

	var body struct {
		ShortURL string `json:"short_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	parts := strings.Split(body.ShortURL, "/")
	return parts[len(parts)-1]
}

func TestShortenURL(t *testing.T) {
	router := newRouter()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid url", `{"url":"https://go.dev"}`, http.StatusOK},
		{"invalid json", `url=https://go.dev`, http.StatusBadRequest},
		{"empty url", `{"url":""}`, http.StatusBadRequest},
		{"url without scheme", `{"url":"go.dev"}`, http.StatusBadRequest},
		{"non http scheme", `{"url":"javascript:alert(1)"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(router, http.MethodPost, "/shorten", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

func TestShortenURLUsesBaseURL(t *testing.T) {
	t.Setenv("BASE_URL", "https://sho.rt")
	router := newRouter()

	rec := doRequest(router, http.MethodPost, "/shorten", `{"url":"https://go.dev"}`)
	if !strings.Contains(rec.Body.String(), `"short_url":"https://sho.rt/`) {
		t.Fatalf("expected short_url on BASE_URL, got %s", rec.Body)
	}
}

func TestOriginalURL(t *testing.T) {
	router := newRouter()
	code := shorten(t, router, "https://go.dev/doc")

	rec := doRequest(router, http.MethodGet, "/"+code, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if want := `{"originalURL":"https://go.dev/doc"}`; rec.Body.String() != want {
		t.Fatalf("body = %s, want %s", rec.Body, want)
	}
}

func TestRedirectURL(t *testing.T) {
	router := newRouter()
	code := shorten(t, router, "https://go.dev/blog")

	rec := doRequest(router, http.MethodGet, "/redirect/"+code, "")
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "https://go.dev/blog" {
		t.Fatalf("Location = %q", got)
	}
}

func TestUnknownShortURL(t *testing.T) {
	router := newRouter()

	for _, path := range []string{"/zzzzzz", "/redirect/zzzzzz"} {
		rec := doRequest(router, http.MethodGet, path, "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET %s status = %d, want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}

func TestHealth(t *testing.T) {
	rec := doRequest(newRouter(), http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK || rec.Body.String() != `{"status":"ok"}` {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
}
