package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Read env file
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/handlers"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/middleware"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

func main() {
	// Load environment variables from .env file
	// Note: On Render, it will use actual environment variables instead of .env
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ No .env file found, using system environment variables")
	}

	// // Debug db
	// fmt.Println("Current DB URL:", os.Getenv("DATABASE_URL"))

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println("⚠️  Warning: .env file not found, checking system variables...")
	// }

	// dsn := os.Getenv("DATABASE_URL")
	// fmt.Printf("🔍 Debug: Loaded DSN is: [%s]\n", dsn) // Check if it print data in env

	// if dsn == "" {
	// 	log.Fatal("❌ Error: DATABASE_URL is empty. Please check your .env file.")
	// }

	// Initialize Infrastructure
	rand.Seed(time.Now().UnixNano())
	store.InitDB() // Now correctly uses DATABASE_URL from .env or Render

	// App Lifecycle Management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Background Services
	go handlers.MainHub.Run()
	go handlers.StartClickWorker(ctx)

	// Worker for scheduled cleanup
    go func() {
        ticker := time.NewTicker(6 * time.Hour)
        for {
            select {
            case <-ticker.C:
                log.Println(" [System] Running scheduled cleanup...")
                handlers.RunSystemAutoCleanup()
            case <-ctx.Done():
                return
            }
        }
    }()

	// Router Configuration
	r := gin.Default()

	// Load templates and static files
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// --- ROUTES ---

	// Auth (Public)
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Admin (Protected)
	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.html", nil)
	})

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminCheck())
	{
		admin.GET("/users", handlers.GetAllUsers)
		admin.PUT("/users/:id/role", handlers.UpdateUserRole)
		admin.DELETE("/users/:id", handlers.DeleteUser)
	}

	// Core API (Protected)
	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired())
	{
		protected.POST("/shorten", handlers.CreateShortLink)
		protected.GET("/links", handlers.GetMyLinks)
		protected.PUT("/links/:id", handlers.UpdateLink)
		protected.DELETE("/links/:id", handlers.DeleteLink)
		protected.DELETE("/links/cleanup", handlers.CleanupUserExpiredLinks)
	}

	// System Routes
	r.GET("/ws", handlers.HandleWebSocket)
	r.GET("/:code", handlers.RedirectLink)

	// Server Startup & Graceful Shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		fmt.Printf("-------------------------------------------\n")
		fmt.Printf("🚀 Service is running on port %s\n", port)
		fmt.Printf("-------------------------------------------\n")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️  Shutting down gracefully...")

	// Trigger worker cleanup
	cancel()

	// Shutdown HTTP server with 5s timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	time.Sleep(1 * time.Second) // Final wait for workers
	log.Println("✅ Server exited safely.")
}
