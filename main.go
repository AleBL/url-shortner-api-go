package main

import (
	"fmt"
	routes "url-shortener/internal"
	"url-shortener/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	routes.SetupRoutes(router)

	_ = database.ConnectDB()

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Error starting server:", err)
	}
	fmt.Println("Server running on port 8080...")
}
