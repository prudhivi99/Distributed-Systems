package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// User struct - this is our data model
// Think of it like a class in Java or a model in Python
type User struct {
	ID        string    `json:"id"` // `json:"id"` means when we convert to JSON, use "id" as the field name
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// Response struct - standard API response format
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"` // omitempty means: don't include if empty
	Data    interface{} `json:"data,omitempty"`    // interface{} means: can be any type
	Error   string      `json:"error,omitempty"`
}

// In-memory storage (like a HashMap/Dictionary)
// Later we'll replace this with MongoDB
var users = make(map[string]User)

func main() {
	// Create a new router - this handles all HTTP requests
	router := mux.NewRouter()

	// Register endpoints (routes)
	// Think of this like: "When someone visits /health, call the HealthCheck function"
	router.HandleFunc("/health", HealthCheck).Methods("GET")
	router.HandleFunc("/", HomeHandler).Methods("GET")
	router.HandleFunc("/api/users", CreateUser).Methods("POST")
	router.HandleFunc("/api/users", GetAllUsers).Methods("GET")
	router.HandleFunc("/api/users/{id}", GetUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", UpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", DeleteUser).Methods("DELETE")

	// Start the server
	port := "8080"
	fmt.Printf("🚀 Server starting on port %s...\n", port)
	fmt.Printf("📖 API Documentation:\n")
	fmt.Printf("   GET    http://localhost:%s/health\n", port)
	fmt.Printf("   GET    http://localhost:%s/\n", port)
	fmt.Printf("   POST   http://localhost:%s/api/users\n", port)
	fmt.Printf("   GET    http://localhost:%s/api/users\n", port)
	fmt.Printf("   GET    http://localhost:%s/api/users/{id}\n", port)
	fmt.Printf("   PUT    http://localhost:%s/api/users/{id}\n", port)
	fmt.Printf("   DELETE http://localhost:%s/api/users/{id}\n", port)

	// This blocks and keeps the server running
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// ============================================
// HANDLER FUNCTIONS
// ============================================

// HealthCheck - returns if the service is running
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Service is healthy",
		Data: map[string]string{
			"status": "UP",
		},
	})
}

// HomeHandler - returns API information
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Welcome to User Management API",
		Data: map[string]interface{}{
			"version":     "1.0.0",
			"total_users": len(users),
		},
	})
}

// CreateUser - creates a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User

	// Decode JSON from request body into user struct
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate input
	if user.Name == "" || user.Email == "" {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Name and Email are required",
		})
		return
	}

	// Generate a simple ID (in real app, MongoDB will generate this)
	user.ID = fmt.Sprintf("user_%d", time.Now().UnixNano())
	user.CreatedAt = time.Now()

	// Store in our "database" (map)
	users[user.ID] = user

	respondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

// GetAllUsers - retrieves all users
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Convert map to slice (array)
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d users", len(userList)),
		Data:    userList,
	})
}

// GetUser - retrieves a single user by ID
func GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract {id} from URL path
	vars := mux.Vars(r)
	id := vars["id"]

	// Look up user in our map
	user, exists := users[id]
	if !exists {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateUser - updates an existing user
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if user exists
	existingUser, exists := users[id]
	if !exists {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	// Decode new user data
	var updatedUser User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Keep original ID and CreatedAt
	updatedUser.ID = existingUser.ID
	updatedUser.CreatedAt = existingUser.CreatedAt

	// Update in map
	users[id] = updatedUser

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "User updated successfully",
		Data:    updatedUser,
	})
}

// DeleteUser - deletes a user by ID
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if user exists
	if _, exists := users[id]; !exists {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	// Delete from map
	delete(users, id)

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// respondJSON - helper to send JSON responses
func respondJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
