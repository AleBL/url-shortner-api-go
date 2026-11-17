package routes

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"url-shortener/controllers"
	"url-shortener/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

const defaultRateLimitPerMinute = 10

func SetupRoutes(router *gin.Engine) {
	configureClientIP(router)

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
	router.POST("/shorten", middleware.RateLimit(rateLimitPerMinute()), controllers.ShortenURL)
	// The short link returned by /shorten redirects; the JSON lookup lives under /api.
	router.GET("/:shortURL", controllers.RedirectURL)
	router.GET("/api/urls/:shortURL", controllers.OriginalURL)

	if os.Getenv("DEBUG_CLIENT_IP") == "true" {
		router.GET("/debug/client-ip", debugClientIP)
	}
}

// configureClientIP decides where c.ClientIP() comes from. X-Forwarded-For is
// never trusted because clients can forge it. Behind a proxy that puts the client
// address in a header the client cannot override (CLIENT_IP_HEADER), that header
// is used; otherwise the TCP peer address is.
func configureClientIP(router *gin.Engine) {
	router.TrustedPlatform = os.Getenv("CLIENT_IP_HEADER")
	_ = router.SetTrustedProxies(nil)
}

// rateLimitPerMinute reads RATE_LIMIT_PER_MINUTE for POST /shorten; 0 disables it.
func rateLimitPerMinute() int {
	value := os.Getenv("RATE_LIMIT_PER_MINUTE")
	if value == "" {
		return defaultRateLimitPerMinute
	}

	perMinute, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Invalid RATE_LIMIT_PER_MINUTE %q, using %d", value, defaultRateLimitPerMinute)
		return defaultRateLimitPerMinute
	}
	return perMinute
}

// debugClientIP shows the caller's IP-related headers, to find out which header
// the hosting platform sets (see CLIENT_IP_HEADER). Enabled by DEBUG_CLIENT_IP=true.
func debugClientIP(c *gin.Context) {
	headers := gin.H{}
	for _, name := range []string{"X-Forwarded-For", "X-Real-Ip", "True-Client-Ip", "Cf-Connecting-Ip"} {
		headers[name] = c.GetHeader(name)
	}
	c.JSON(http.StatusOK, gin.H{"client_ip": c.ClientIP(), "remote_addr": c.Request.RemoteAddr, "headers": headers})
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
