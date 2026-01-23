package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/handlers"
)

func main() {
	// Connect to database
	database, err := db.NewPostgresDB("localhost", 5432, "minisys", "minisys123", "minisys")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create repository and handler
	productRepo := db.NewProductRepository(database)
	productHandler := handlers.NewProductHandler(productRepo)

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
