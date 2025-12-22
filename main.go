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
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/handlers"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/middleware"
	"github.com/tuannho0802/URL-Shortener-Service-Golang-/store"
)

func main() {
	
	rand.Seed(time.Now().UnixNano())
	store.InitDB()

	// Initialize a Context to manage the app lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run background progress
	go handlers.MainHub.Run()
	// Start Click Worker
	go handlers.StartClickWorker(ctx)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// AUTH ROUTES Public
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// PROTECTED ROUTES Private

	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired())
	{
		protected.POST("/shorten", handlers.CreateShortLink)
		protected.GET("/links", handlers.GetMyLinks)
		protected.PUT("/links/:id", handlers.UpdateLink)
		protected.DELETE("/links/:id", handlers.DeleteLink)

	}

	r.GET("/ws", handlers.HandleWebSocket)

	r.GET("/:code", handlers.RedirectLink)

	// Config HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Run the server within a Goroutine to avoid clogging the main thread
	go func() {
		fmt.Print("-------------------------------------------\n")
		fmt.Println("🚀 URL Shortener Service is running on http://localhost:8080")
		fmt.Print("-------------------------------------------\n")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// Listen interrupt from OS
	quit := make(chan os.Signal, 1)
	// SIGINT: Ctrl+C, SIGTERM: Shutdown signal from Docker/System
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Chờ ở đây cho đến khi có tín hiệu

	log.Println("⚠️  Đang bắt đầu quá trình tắt an toàn...")

	// Run a Cancel command to have the Worker save the last click.
	cancel()

	// Custom for shutdown server gracefully 5s
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Wait a little bit fot the processRemainingClicks run to finish
	time.Sleep(1 * time.Second)
	log.Println("✅ Server đã tắt hoàn toàn. Dữ liệu đã được bảo vệ.")
}
