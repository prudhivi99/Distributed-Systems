package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/client"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/handlers"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/messaging"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/publisher"
)

func main() {
	// Connect to PostgreSQL
	database, err := db.NewPostgresDB("localhost", 5432, "minisys", "minisys123", "minisys")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Connect to RabbitMQ
	rabbitMQ, err := messaging.NewRabbitMQ("localhost", 5672, "guest", "guest")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	// Create publisher
	orderPublisher, err := publisher.NewOrderPublisher(rabbitMQ)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}

	// Create Product Service client (HTTP)
	productClient := client.NewProductClient("http://localhost:8081")

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
	log.Println("🚀 Order Service starting on http://localhost:8082")
	log.Println("   Publishing events to RabbitMQ")
	router.Run(":8082")
}
