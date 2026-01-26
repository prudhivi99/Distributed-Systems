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
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/config"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/discovery"
)

type Gateway struct {
	consul   *discovery.ConsulClient
	proxies  map[string]*httputil.ReverseProxy
	mutex    sync.RWMutex
	services map[string]string
}

func NewGateway(consul *discovery.ConsulClient) *Gateway {
	g := &Gateway{
		consul:   consul,
		proxies:  make(map[string]*httputil.ReverseProxy),
		services: make(map[string]string),
	}

	g.discoverServices()
	go g.watchServices()

	return g
}

func (g *Gateway) discoverServices() {
	services := []string{"product-service", "order-service"}

	for _, svc := range services {
		url, err := g.consul.GetServiceURL(svc)
		if err != nil {
			log.Printf("⚠️ Service %s not found: %v", svc, err)
			// Use K8s DNS as fallback
			switch svc {
			case "product-service":
				url = "http://product-service:8081"
			case "order-service":
				url = "http://order-service:8082"
			}
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

func (g *Gateway) ProxyProducts(c *gin.Context) {
	proxy := g.getProxy("product-service")
	if proxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "product-service unavailable"})
		return
	}
	log.Printf("🔀 Routing %s %s → product-service", c.Request.Method, c.Request.URL.Path)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (g *Gateway) ProxyOrders(c *gin.Context) {
	proxy := g.getProxy("order-service")
	if proxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "order-service unavailable"})
		return
	}
	log.Printf("🔀 Routing %s %s → order-service", c.Request.Method, c.Request.URL.Path)
	proxy.ServeHTTP(c.Writer, c.Request)
}

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

func (g *Gateway) ListServices(c *gin.Context) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{"services": g.services})
}

func main() {
	cfg := config.Load()

	consul, err := discovery.NewConsulClient(cfg.ConsulHost, cfg.ConsulPort)
	if err != nil {
		log.Printf("⚠️ Failed to connect to Consul, using K8s DNS: %v", err)
	}

	gateway := NewGateway(consul)

	router := gin.Default()

	router.GET("/health", gateway.HealthCheck)
	router.GET("/services", gateway.ListServices)

	router.Any("/products", gateway.ProxyProducts)
	router.Any("/products/*path", gateway.ProxyProducts)
	router.Any("/orders", gateway.ProxyOrders)
	router.Any("/orders/*path", gateway.ProxyOrders)

	log.Println("🚀 API Gateway starting on http://0.0.0.0:8080")
	router.Run(":8080")
}
