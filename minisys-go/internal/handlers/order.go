package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/client"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
)

type OrderHandler struct {
	repo          *db.OrderRepository
	productClient *client.ProductClient
}

func NewOrderHandler(repo *db.OrderRepository, productClient *client.ProductClient) *OrderHandler {
	return &OrderHandler{
		repo:          repo,
		productClient: productClient,
	}
}

// HealthCheck returns server status
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "order-service"})
}

// ListOrders returns all orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	orders, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrder returns a single order with items
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build order with product details from Product Service
	order := models.Order{
		CustomerName: req.CustomerName,
		Status:       "pending",
	}

	var totalAmount float64

	for _, item := range req.Items {
		// Call Product Service to get product details
		log.Printf("📞 Calling Product Service for product %d", item.ProductID)
		product, err := h.productClient.GetProduct(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderItem := models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       product.Price,
		}

		totalAmount += product.Price * float64(item.Quantity)
		order.Items = append(order.Items, orderItem)
	}

	order.TotalAmount = totalAmount

	// Save to database
	if err := h.repo.Create(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Order %d created with total $%.2f", order.ID, order.TotalAmount)
	c.JSON(http.StatusCreated, order)
}

// UpdateOrderStatus updates the order status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if err := h.repo.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated"})
}
