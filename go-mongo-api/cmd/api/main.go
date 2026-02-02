package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prudhivi99/Distributed-Systems/go-mongo-api/internal/database"
	"github.com/prudhivi99/Distributed-Systems/go-mongo-api/internal/handlers"
	"github.com/prudhivi99/Distributed-Systems/go-mongo-api/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	db, err := database.Connect(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Disconnect()

	// Create handler with database dependency
	userHandler := handlers.NewUserHandler(db)

	// Create router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/", userHandler.Home).Methods("GET")
	router.HandleFunc("/health", userHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/api/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Print startup information
	fmt.Println("🚀 User Management API - Clean Architecture")
	fmt.Println("==========================================")
	fmt.Printf("📍 Server starting on port %s\n", cfg.Port)
	fmt.Println("==========================================")
	fmt.Println("📖 API Endpoints:")
	fmt.Printf("   GET    http://localhost:%s/\n", cfg.Port)
	fmt.Printf("   GET    http://localhost:%s/health\n", cfg.Port)
	fmt.Printf("   POST   http://localhost:%s/api/users\n", cfg.Port)
	fmt.Printf("   GET    http://localhost:%s/api/users\n", cfg.Port)
	fmt.Printf("   GET    http://localhost:%s/api/users/{id}\n", cfg.Port)
	fmt.Printf("   PUT    http://localhost:%s/api/users/{id}\n", cfg.Port)
	fmt.Printf("   DELETE http://localhost:%s/api/users/{id}\n", cfg.Port)
	fmt.Println("==========================================")
	fmt.Println("✅ Using MongoDB with Clean Architecture")
	fmt.Println("==========================================")

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
