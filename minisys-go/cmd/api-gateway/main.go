package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/discovery"
)

// Gateway handles request routing with service discovery
type Gateway struct {
	consul   *discovery.ConsulClient
	proxies  map[string]*httputil.ReverseProxy
	mutex    sync.RWMutex
	services map[string]string // service name -> current URL
}

// NewGateway creates a new API Gateway with Consul integration
func NewGateway(consul *discovery.ConsulClient) *Gateway {
	g := &Gateway{
		consul:   consul,
		proxies:  make(map[string]*httputil.ReverseProxy),
		services: make(map[string]string),
	}

	// Initial service discovery
	g.discoverServices()

	// Watch for service changes
	go g.watchServices()

	return g
}

func (g *Gateway) discoverServices() {
	services := []string{"product-service", "order-service"}

	for _, svc := range services {
		url, err := g.consul.GetServiceURL(svc)
		if err != nil {
			log.Printf("⚠️ Service %s not found: %v", svc, err)
			continue
		}

		g.updateProxy(svc, url)
	}
}

func (g *Gateway) updateProxy(serviceName, serviceURL string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	target, err := url.Parse(serviceURL)
	if err != nil {
		log.Printf("❌ Invalid URL for %s: %v", serviceName, err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("❌ Proxy error for %s: %v", serviceName, err)
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, `{"error": "service unavailable"}`)
	}

	g.proxies[serviceName] = proxy
	g.services[serviceName] = serviceURL
	log.Printf("✅ Updated route: %s → %s", serviceName, serviceURL)
}

func (g *Gateway) watchServices() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		g.discoverServices()
	}
}

func (g *Gateway) getProxy(serviceName string) *httputil.ReverseProxy {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.proxies[serviceName]
}

// ProxyProducts forwards requests to product-service
func (g *Gateway) ProxyProducts(c *gin.Context) {
	proxy := g.getProxy("product-service")
	if proxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product-service unavailable"})
		return
	}

	log.Printf("🔀 Routing %s %s → product-service", c.Request.Method, c.Request.URL.Path)
	proxy.ServeHTTP(c.Writer, c.Request)
}

// ProxyOrders forwards requests to order-service
func (g *Gateway) ProxyOrders(c *gin.Context) {
	proxy := g.getProxy("order-service")
	if proxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "order-service unavailable"})
		return
	}

	log.Printf("🔀 Routing %s %s → order-service", c.Request.Method, c.Request.URL.Path)
	proxy.ServeHTTP(c.Writer, c.Request)
}

// HealthCheck returns gateway and services status
func (g *Gateway) HealthCheck(c *gin.Context) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	statuses := make(map[string]string)
	allHealthy := true

	client := &http.Client{Timeout: 2 * time.Second}

	for name, url := range g.services {
		resp, err := client.Get(url + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			statuses[name] = "unhealthy"
			allHealthy = false
		} else {
			statuses[name] = "healthy"
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

// ListServices returns all discovered services
func (g *Gateway) ListServices(c *gin.Context) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"services": g.services,
	})
}

func main() {
	// Connect to Consul
	consul, err := discovery.NewConsulClient("localhost", 8500)
	if err != nil {
		log.Fatalf("Failed to connect to Consul: %v", err)
	}

	// Create gateway
	gateway := NewGateway(consul)

	// Setup router
	router := gin.Default()

	// Gateway endpoints
	router.GET("/health", gateway.HealthCheck)
	router.GET("/services", gateway.ListServices)

	// Proxy routes
	router.Any("/products", gateway.ProxyProducts)
	router.Any("/products/*path", gateway.ProxyProducts)
	router.Any("/orders", gateway.ProxyOrders)
	router.Any("/orders/*path", gateway.ProxyOrders)

	// Start gateway
	log.Println("🚀 API Gateway starting on http://localhost:8080")
	log.Println("   Using Consul for service discovery")
	router.Run(":8080")
}
