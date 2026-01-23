package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/client"
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

	// Create Product Service client
	productClient := client.NewProductClient("http://localhost:8080")

	// Create repository and handler
	orderRepo := db.NewOrderRepository(database)
	orderHandler := handlers.NewOrderHandler(orderRepo, productClient)

	// Setup router
	router := gin.Default()

	// Register routes
	router.GET("/health", orderHandler.HealthCheck)
	router.GET("/orders", orderHandler.ListOrders)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.POST("/orders", orderHandler.CreateOrder)
	router.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)

	// Start server on port 8081
	log.Println("🚀 Order Service starting on http://localhost:8081")
	router.Run(":8081")
}
