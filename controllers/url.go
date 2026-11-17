package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"url-shortener/internal/logic"
	"url-shortener/internal/services"

	"github.com/gin-gonic/gin"
)

// maxShortURLAttempts bounds retries when a generated code already exists.
const maxShortURLAttempts = 5

func ShortenURL(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading JSON"})
		return
	}

	originalURL := strings.TrimSpace(request.URL)

	if originalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL not provided"})
		return
	}

	if err := logic.ValidateURL(originalURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL, err := saveWithUniqueShortURL(originalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving URL"})
		return
	}

	// Platforms like Render terminate TLS at their proxy and forward plain HTTP.
	isHTTPS := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.JSON(http.StatusOK, gin.H{"short_url": logic.GetFullURL(shortURL, os.Getenv("BASE_URL"), isHTTPS, c.Request.Host)})
}

func saveWithUniqueShortURL(originalURL string) (string, error) {
	for attempt := 0; attempt < maxShortURLAttempts; attempt++ {
		shortURL := logic.GenerateShortURL()

		err := services.SaveURL(originalURL, shortURL)
		if err == nil {
			return shortURL, nil
		}
		if !errors.Is(err, services.ErrShortURLTaken) {
			return "", err
		}
	}
	return "", services.ErrShortURLTaken
}

// getOriginalURL writes the error response itself, so callers only return on error.
func getOriginalURL(c *gin.Context) (string, bool) {
	shortURL := c.Param("shortURL")

	if shortURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short URL"})
		return "", false
	}

	originalURL, err := services.FindUrlByShortUrl(shortURL)
	if errors.Is(err, services.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return "", false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching URL"})
		return "", false
	}

	return originalURL, true
}

func RedirectURL(c *gin.Context) {
	originalURL, ok := getOriginalURL(c)
	if !ok {
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

func OriginalURL(c *gin.Context) {
	originalURL, ok := getOriginalURL(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{"originalURL": originalURL})
}
