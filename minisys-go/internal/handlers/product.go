package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
)

type ProductHandler struct {
	repo *db.ProductRepository
}

func NewProductHandler(repo *db.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// HealthCheck returns server status
func (h *ProductHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// ListProducts returns all products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct returns a single product
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// CreateProduct adds a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.repo.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// DeleteProduct removes a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
