package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/cache"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/consumer"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/discovery"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/handlers"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/messaging"
)

const (
	serviceName = "product-service"
	serviceID   = "product-service-1"
	servicePort = 8081
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

	// Connect to Consul
	consul, err := discovery.NewConsulClient("localhost", 8500)
	if err != nil {
		log.Fatalf("Failed to connect to Consul: %v", err)
	}

	// Register with Consul
	err = consul.Register(discovery.ServiceConfig{
		Name: serviceName,
		ID:   serviceID,
		Port: servicePort,
		Tags: []string{"api", "products"},
	})
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// Deregister on shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		consul.Deregister(serviceID)
		os.Exit(0)
	}()

	// Create repositories
	productRepo := db.NewProductRepository(database)
	cachedRepo := db.NewCachedProductRepository(productRepo, redisCache)

	// Create handler
	productHandler := handlers.NewProductHandler(cachedRepo)

	// Start event consumer
	go startEventConsumer(rabbitMQ, productRepo, redisCache)

	// Setup router
	router := gin.Default()

	router.GET("/health", productHandler.HealthCheck)
	router.GET("/products", productHandler.ListProducts)
	router.GET("/products/:id", productHandler.GetProduct)
	router.POST("/products", productHandler.CreateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	// Start server
	log.Printf("🚀 %s starting on http://localhost:%d", serviceName, servicePort)
	log.Println("   Registered with Consul")
	router.Run(":8081")
}

func startEventConsumer(mq *messaging.RabbitMQ, repo *db.ProductRepository, cache *cache.RedisCache) {
	if err := mq.DeclareQueue("order.created"); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	messages, err := mq.Consume("order.created")
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	inventoryConsumer := consumer.NewInventoryConsumer(repo, cache)
	inventoryConsumer.ProcessOrderCreated(messages)
}
