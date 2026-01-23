package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceConfig holds service routing configuration
type ServiceConfig struct {
	Name   string
	URL    string
	Prefix string
}

// Gateway handles request routing
type Gateway struct {
	services map[string]*httputil.ReverseProxy
	configs  []ServiceConfig
}

// NewGateway creates a new API Gateway
func NewGateway(configs []ServiceConfig) *Gateway {
	g := &Gateway{
		services: make(map[string]*httputil.ReverseProxy),
		configs:  configs,
	}

	for _, cfg := range configs {
		target, err := url.Parse(cfg.URL)
		if err != nil {
			log.Fatalf("Invalid URL for service %s: %v", cfg.Name, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		// Custom error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("❌ Proxy error for %s: %v", cfg.Name, err)
			w.WriteHeader(http.StatusBadGateway)
			io.WriteString(w, `{"error": "service unavailable"}`)
		}

		g.services[cfg.Prefix] = proxy
		log.Printf("✅ Registered route %s/* → %s (%s)", cfg.Prefix, cfg.URL, cfg.Name)
	}

	return g
}

// ProxyRequest forwards request to appropriate service
func (g *Gateway) ProxyRequest(c *gin.Context) {
	path := c.Request.URL.Path

	// Find matching service
	for _, cfg := range g.configs {
		if len(path) >= len(cfg.Prefix) && path[:len(cfg.Prefix)] == cfg.Prefix {
			log.Printf("🔀 Routing %s %s → %s", c.Request.Method, path, cfg.Name)

			proxy := g.services[cfg.Prefix]
			proxy.ServeHTTP(c.Writer, c.Request)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
}

// HealthCheck returns gateway status
func (g *Gateway) HealthCheck(c *gin.Context) {
	// Check all services
	statuses := make(map[string]string)
	allHealthy := true

	client := &http.Client{Timeout: 2 * time.Second}

	for _, cfg := range g.configs {
		resp, err := client.Get(cfg.URL + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			statuses[cfg.Name] = "unhealthy"
			allHealthy = false
		} else {
			statuses[cfg.Name] = "healthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	status := "healthy"
	if !allHealthy {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"service":  "api-gateway",
		"services": statuses,
	})
}

func main() {
	// Service configurations
	services := []ServiceConfig{
		{
			Name:   "product-service",
			URL:    "http://localhost:8081",
			Prefix: "/products",
		},
		{
			Name:   "order-service",
			URL:    "http://localhost:8082",
			Prefix: "/orders",
		},
	}

	// Create gateway
	gateway := NewGateway(services)

	// Setup router
	router := gin.Default()

	// Gateway health check
	router.GET("/health", gateway.HealthCheck)

	// Proxy all other requests
	router.Any("/products", gateway.ProxyRequest)
	router.Any("/products/*path", gateway.ProxyRequest)
	router.Any("/orders", gateway.ProxyRequest)
	router.Any("/orders/*path", gateway.ProxyRequest)

	// Start gateway on port 8080
	log.Println("🚀 API Gateway starting on http://localhost:8080")
	log.Println("   Routes:")
	log.Println("   /products/* → Product Service (8081)")
	log.Println("   /orders/*   → Order Service (8082)")
	router.Run(":8080")
}
