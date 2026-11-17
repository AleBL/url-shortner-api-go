package routes

import (
	"net/http"
	"os"
	"strings"
	"time"
	"url-shortener/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func SetupRoutes(router *gin.Engine) {
	// Must run before the routes are registered to apply to them.
	if origins := allowedOrigins(); len(origins) > 0 {
		router.Use(cors.New(cors.Config{
			AllowOrigins: origins,
			AllowMethods: []string{http.MethodGet, http.MethodPost},
			AllowHeaders: []string{"Content-Type"},
			MaxAge:       12 * time.Hour,
		}))
	}

	router.GET("/health", Health)
	router.POST("/shorten", controllers.ShortenURL)
	// The short link returned by /shorten redirects; the JSON lookup lives under /api.
	router.GET("/:shortURL", controllers.RedirectURL)
	router.GET("/api/urls/:shortURL", controllers.OriginalURL)
}

// allowedOrigins reads ALLOWED_ORIGINS, a comma-separated list of frontend
// origins such as "https://app.example.com,http://localhost:5173". When empty,
// browsers only accept same-origin calls.
func allowedOrigins() []string {
	var origins []string
	for _, origin := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, strings.TrimRight(origin, "/"))
		}
	}
	return origins
}
