package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/client"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/config"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/discovery"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/handlers"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/messaging"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/publisher"
)

const (
	serviceName = "order-service"
	serviceID   = "order-service-1"
)

func main() {
	// Load configuration
	cfg := config.Load()
	servicePort := 8082

	// Connect to PostgreSQL
	database, err := db.NewPostgresDB(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Connect to RabbitMQ
	rabbitMQ, err := messaging.NewRabbitMQ(cfg.RabbitMQHost, cfg.RabbitMQPort, cfg.RabbitMQUser, cfg.RabbitMQPassword)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	// Connect to Consul
	consul, err := discovery.NewConsulClient(cfg.ConsulHost, cfg.ConsulPort)
	if err != nil {
		log.Fatalf("Failed to connect to Consul: %v", err)
	}

	// Register with Consul
	err = consul.Register(discovery.ServiceConfig{
		Name: serviceName,
		ID:   serviceID,
		Port: servicePort,
		Tags: []string{"api", "orders"},
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

	// Discover Product Service from Consul
	productServiceURL, err := consul.GetServiceURL("product-service")
	if err != nil {
		log.Printf("⚠️ Product service not found, using default: %v", err)
		productServiceURL = "http://product-service:8081"
	}
	log.Printf("📍 Discovered product-service at: %s", productServiceURL)

	// Create publisher
	orderPublisher, err := publisher.NewOrderPublisher(rabbitMQ)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}

	// Create Product Service client
	productClient := client.NewProductClient(productServiceURL)

	// Create repository and handler
	orderRepo := db.NewOrderRepository(database)
	orderHandler := handlers.NewOrderHandler(orderRepo, productClient, orderPublisher)

	// Setup router
	router := gin.Default()

	router.GET("/health", orderHandler.HealthCheck)
	router.GET("/orders", orderHandler.ListOrders)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.POST("/orders", orderHandler.CreateOrder)
	router.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)

	// Start server
	log.Printf("🚀 %s starting on http://0.0.0.0:%d", serviceName, servicePort)
	router.Run(":8082")
}
