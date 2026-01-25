package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/cache"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/consumer"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/handlers"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/messaging"
)

func main() {
	// Connect to PostgreSQL
	database, err := db.NewPostgresDB("localhost", 5432, "minisys", "minisys123", "minisys")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Connect to Redis
	redisCache, err := cache.NewRedisCache("localhost", 6379, 5*time.Minute)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Connect to RabbitMQ
	rabbitMQ, err := messaging.NewRabbitMQ("localhost", 5672, "guest", "guest")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	// Create repositories
	productRepo := db.NewProductRepository(database)
	cachedRepo := db.NewCachedProductRepository(productRepo, redisCache)

	// Create handler
	productHandler := handlers.NewProductHandler(cachedRepo)

	// Start event consumer in goroutine (pass cache for invalidation)
	go startEventConsumer(rabbitMQ, productRepo, redisCache)

	// Setup router
	router := gin.Default()

	router.GET("/health", productHandler.HealthCheck)
	router.GET("/products", productHandler.ListProducts)
	router.GET("/products/:id", productHandler.GetProduct)
	router.POST("/products", productHandler.CreateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	// Start server
	log.Println("🚀 Product Service starting on http://localhost:8081")
	log.Println("   Consuming events from RabbitMQ")
	router.Run(":8081")
}

func startEventConsumer(mq *messaging.RabbitMQ, repo *db.ProductRepository, cache *cache.RedisCache) {
	// Declare queue
	if err := mq.DeclareQueue("order.created"); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Start consuming
	messages, err := mq.Consume("order.created")
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	// Process messages (pass cache for invalidation)
	inventoryConsumer := consumer.NewInventoryConsumer(repo, cache)
	inventoryConsumer.ProcessOrderCreated(messages)
}
