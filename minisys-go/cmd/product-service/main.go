package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/cache"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/handlers"
)

func main() {
	// Connect to PostgreSQL
	database, err := db.NewPostgresDB("localhost", 5432, "minisys", "minisys123", "minisys")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Connect to Redis (5 minute TTL)
	redisCache, err := cache.NewRedisCache("localhost", 6379, 5*time.Minute)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Create repositories
	productRepo := db.NewProductRepository(database)
	cachedRepo := db.NewCachedProductRepository(productRepo, redisCache)

	// Create handler
	productHandler := handlers.NewProductHandler(cachedRepo)

	// Setup router
	router := gin.Default()

	// Register routes
	router.GET("/health", productHandler.HealthCheck)
	router.GET("/products", productHandler.ListProducts)
	router.GET("/products/:id", productHandler.GetProduct)
	router.POST("/products", productHandler.CreateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	// Start server
	log.Println("🚀 Product Service starting on http://localhost:8080")
	router.Run(":8080")
}
