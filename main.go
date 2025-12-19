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

	// Hub manage socket
	go handlers.MainHub.Run()

	r := gin.Default()
	// Connect the frontend
	r.Static("/static", "./static")

	// Route page return UI
	r.GET("/", func(c *gin.Context) {

		c.File("static/index.html")
	})

	// Define Routes API
	r.POST("/shorten", handlers.CreateShortLink) // Create Link

	r.GET("/links", handlers.GetAllLinks) // Get Link List

	r.PUT("/links/:id", handlers.UpdateLink)
	r.DELETE("/links/:id", handlers.DeleteLink)

	r.GET("/ws", handlers.HandleWebSocket)

	r.GET("/:code", handlers.RedirectLink) // Redirect Link

	r.Run(":8080") // Run server on port 8080
}
