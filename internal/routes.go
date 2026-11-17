package routes

import (
	"net/http"
	"url-shortener/controllers"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func SetupRoutes(router *gin.Engine) {
	router.GET("/health", Health)
	router.POST("/shorten", controllers.ShortenURL)
	// The short link returned by /shorten redirects; the JSON lookup lives under /api.
	router.GET("/:shortURL", controllers.RedirectURL)
	router.GET("/api/urls/:shortURL", controllers.OriginalURL)
}
