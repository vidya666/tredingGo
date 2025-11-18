package main

import (
	"log"
	"math/rand"
	"time"
	"trading-go/handlers"
	"trading-go/middleware"
	"trading-go/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Start background services
	go services.UpdatePrices()
	go handlers.BroadcastPrices()

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Public routes (no authentication required)
	public := r.Group("/api")
	{
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
		public.GET("/prices", handlers.GetPrices) // Prices are public
	}

	// Protected routes (authentication required)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", handlers.GetProfile)
		protected.POST("/orders", handlers.CreateOrder)
		protected.GET("/orders", handlers.GetOrders)
	}

	// WebSocket endpoint (public for real-time price updates)
	r.GET("/ws", handlers.HandleWebSocket)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	log.Println("========================================")
	log.Println("Trading Dashboard Server")
	log.Println("========================================")
	log.Println("Server starting on :8080")
	log.Println("Demo Account:")
	log.Println("  Username: demo")
	log.Println("  Password: demo123")
	log.Println("========================================")

	r.Run(":8080")
}
