package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/handlers"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"

	"math/rand"
	"time"
)

func main() {
	// Seed for random func
	rand.Seed(time.Now().UnixNano())

	// Create db
	store.InitDB()

	r := gin.Default()

	// Define Routes API
	r.POST("/shorten", handlers.CreateShortLink) // Create Link
	r.GET("/:code", handlers.RedirectLink)       // Redirect Link
	r.GET("/links", handlers.GetAllLinks)        // Get Link List

	r.Run(":8080") // Run server on port 8080
}
