package logic

import (
	"strings"
	"testing"
)

func TestGenerateShortURL(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		code := GenerateShortURL()
		if len(code) != shortURLLength {
			t.Fatalf("expected length %d, got %q", shortURLLength, code)
		}
		for _, char := range code {
			if !strings.ContainsRune(shortURLChars, char) {
				t.Fatalf("unexpected character %q in %q", char, code)
			}
		}
		seen[code] = true
	}
	if len(seen) < 990 {
		t.Fatalf("expected mostly unique codes, got %d distinct of 1000", len(seen))
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https url", "https://go.dev/doc", false},
		{"http url with query", "http://example.com/path?q=1", false},
		{"missing scheme", "example.com", true},
		{"plain text", "not a url", true},
		{"relative path", "/local/path", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"ftp scheme", "ftp://example.com/file", true},
		{"scheme without host", "https://", true},
		{"too long", "https://example.com/" + strings.Repeat("a", maxURLLength), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestGetFullURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		hasTLS  bool
		host    string
		want    string
	}{
		{"request without TLS", "", false, "localhost:8080", "http://localhost:8080/abc123"},
		{"request with TLS", "", true, "sho.rt", "https://sho.rt/abc123"},
		{"base URL wins over request", "https://sho.rt", false, "internal:8080", "https://sho.rt/abc123"},
		{"base URL trailing slash", "https://sho.rt/", false, "internal:8080", "https://sho.rt/abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFullURL("abc123", tt.baseURL, tt.hasTLS, tt.host)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
