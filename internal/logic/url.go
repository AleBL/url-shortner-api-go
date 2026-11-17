package logic

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strings"
)

const (
	shortURLChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortURLLength = 6
	maxURLLength   = 2048
)

var ErrInvalidURL = errors.New("URL must be an absolute http or https URL")

// GenerateShortURL uses the auto-seeded global generator, which is safe for
// concurrent use; reseeding with the clock made simultaneous requests collide.
func GenerateShortURL() string {
	var shortURL strings.Builder
	for i := 0; i < shortURLLength; i++ {
		shortURL.WriteByte(shortURLChars[rand.IntN(len(shortURLChars))])
	}
	return shortURL.String()
}

// ValidateURL accepts only absolute http(s) URLs, so the redirect never points
// to a relative path or a non-web scheme.
func ValidateURL(rawURL string) error {
	if len(rawURL) > maxURLLength {
		return ErrInvalidURL
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return ErrInvalidURL
	}
	return nil
}

// GetFullURL builds the public short link. When baseURL is set (BASE_URL env) it
// wins over the request host and scheme.
func GetFullURL(shortURL, baseURL string, isHTTPS bool, host string) string {
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + "/" + shortURL
	}

	scheme := "http"

	if isHTTPS {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s", scheme, host, shortURL)
}
