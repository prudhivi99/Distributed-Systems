package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================================
// DATA MODEL
// ============================================

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// In-memory storage (we will replace with database later )
var products = []Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Quantity: 10},
	{ID: 2, Name: "Mouse", Price: 29.99, Quantity: 100},
	{ID: 3, Name: "Keyboard", Price: 79.99, Quantity: 50},
}

var nextID = 4

// ============================================
// HANDLERS
// ============================================

// HealthCheck returns server health status
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "helthy"})
}

// ListProducts returns all products
func ListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, products)
}

// GetProduct returns a single product by ID
func GetProduct(c *gin.Context) {
	id := c.Param("id")

	for _, p := range products {
		if fmt.Sprintf("%d", p.ID) == id {
			c.JSON(http.StatusOK, p)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
}

// CreateProduct adds a new product
func CreateProduct(c *gin.Context) {

	var newProduct Product

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// Assign ID
	newProduct.ID = nextID
	nextID++

	// Add to slice
	products = append(products, newProduct)

	c.JSON(http.StatusCreated, newProduct)
}

// DeleteProduct removes a product by ID
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	for i, p := range products {
		if fmt.Sprintf("%d", p.ID) == id {
			//Remove from slice
			products = append(products[:i], products[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
}

func main() {
	// Create Gin router
	router := gin.Default()

	// Register routes
	router.GET("/health", HealthCheck)
	router.GET("/products", ListProducts)
	router.GET("/products/:id", GetProduct)
	router.POST("/products", CreateProduct)
	router.DELETE("/products/:id", DeleteProduct)

	// Start server
	fmt.Println("Product Service starting on http://localhost:8080")
	router.Run(":8080")
}
