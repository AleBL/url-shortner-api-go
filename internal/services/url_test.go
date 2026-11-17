package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"url-shortener/internal/database"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "url-shortener-services")
	if err != nil {
		panic(err)
	}
	os.Setenv("DB_PATH", filepath.Join(dir, "test.db"))

	code := m.Run()

	database.Close()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestSaveAndFindURL(t *testing.T) {
	if err := SaveURL("https://go.dev", "svc001"); err != nil {
		t.Fatalf("SaveURL: %v", err)
	}

	got, err := FindUrlByShortUrl("svc001")
	if err != nil {
		t.Fatalf("FindUrlByShortUrl: %v", err)
	}
	if got != "https://go.dev" {
		t.Fatalf("got %q, want %q", got, "https://go.dev")
	}
}

func TestSaveURLDuplicateShortURL(t *testing.T) {
	if err := SaveURL("https://go.dev", "svc002"); err != nil {
		t.Fatalf("first SaveURL: %v", err)
	}

	err := SaveURL("https://example.com", "svc002")
	if !errors.Is(err, ErrShortURLTaken) {
		t.Fatalf("expected ErrShortURLTaken, got %v", err)
	}
}

func TestFindURLNotFound(t *testing.T) {
	_, err := FindUrlByShortUrl("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
