package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Error starting server:", err)
	}
	fmt.Println("Server running on port 8080...")
}
