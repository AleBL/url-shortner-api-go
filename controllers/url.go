package controllers

import (
	"fmt"
	"net/http"
	"os"
	"url-shortener/internal/logic"
	"url-shortener/internal/services"

	"github.com/gin-gonic/gin"
)

func ShortenURL(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading JSON"})
		return
	}

	originalURL := request.URL

	if originalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL not provided"})
		return
	}

	shortURL := logic.GenerateShortURL()

	err := services.URL.SaveURL(originalURL, shortURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving URL"})
		return
	}

	hasTLS := c.Request.TLS != nil
	c.JSON(http.StatusOK, gin.H{"short_url": logic.GetFullURL(shortURL, os.Getenv("BASE_URL"), hasTLS, c.Request.Host)})
}

func getOriginalURL(c *gin.Context) (string, error) {
	shortURL := c.Param("shortURL")

	if shortURL == "" {
		return "", fmt.Errorf("invalid shortURL")
	}

	originalURL, err := services.FindUrlByShortUrl(shortURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return "", err
	}

	return originalURL, nil
}

func RedirectURL(c *gin.Context) {
	originalURL, err := getOriginalURL(c)
	if err != nil {
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

func OriginalURL(c *gin.Context) {
	originalURL, err := getOriginalURL(c)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{"originalURL": originalURL})
}
