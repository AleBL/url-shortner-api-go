package logic

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func GenerateShortURL() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	newRandom := rand.New(rand.NewSource(time.Now().UnixNano()))

	var shortURL strings.Builder
	for i := 0; i < 6; i++ {
		shortURL.WriteByte(chars[newRandom.Intn(len(chars))])
	}
	return shortURL.String()
}

// GetFullURL builds the public short link. When baseURL is set (BASE_URL env) it
// wins over request data, which reports plain http behind a TLS-terminating proxy.
func GetFullURL(shortURL, baseURL string, hasTLS bool, host string) string {
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + "/" + shortURL
	}

	scheme := "http"

	if hasTLS {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s", scheme, host, shortURL)
}
