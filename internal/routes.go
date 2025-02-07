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
	router.GET("/:shortURL", controllers.OriginalURL)
	router.GET("/redirect/:shortURL", controllers.RedirectURL)
}
