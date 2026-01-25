package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/cache"
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

	// Start event consumer in goroutine
	go startEventConsumer(rabbitMQ, productRepo)

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
	err := mq.DeclareQueue("order.created")
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Start consuming messages
	messages, err := mq.Consume("order.created")
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	log.Println("👂 Listening on queue: order.created")

	// Process messages in loop
	for msg := range messages {
		var orderEvent map[string]interface{}
		err := json.Unmarshal(msg.Body, &orderEvent)
		if err != nil {
			log.Printf("❌ Failed to unmarshal event: %v", err)
			msg.Ack(false)
			continue
		}

		log.Printf("📥 Received event: %v", orderEvent)

		// Safely extract fields with type checking
		productIDVal, ok := orderEvent["product_id"]
		if !ok {
			log.Printf("⚠️ Missing product_id in event")
			msg.Ack(false)
			continue
		}

		quantityVal, ok := orderEvent["quantity"]
		if !ok {
			log.Printf("⚠️ Missing quantity in event")
			msg.Ack(false)
			continue
		}

		productID := int(productIDVal.(float64))
		quantity := int(quantityVal.(float64))

		log.Printf("📥 Received order event: product_id=%d, quantity=%d", productID, quantity)

		// Update product quantity in DB
		err = repo.UpdateQuantity(productID, -quantity)
		if err != nil {
			log.Printf("❌ Failed to update quantity: %v", err)
		} else {
			log.Printf("✅ Updated product %d: quantity decreased by %d", productID, quantity)
			// Clear cache for this product
			cache.Delete(fmt.Sprintf("product:%d", productID))
			log.Printf("🗑️ Cleared cache for product %d", productID)
		}

		msg.Ack(false)
	}
}
