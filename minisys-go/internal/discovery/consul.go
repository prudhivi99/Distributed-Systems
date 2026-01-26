package discovery

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *api.Client
}

type ServiceConfig struct {
	Name string
	ID   string
	Port int
	Tags []string
}

func NewConsulClient(host string, port int) (*ConsulClient, error) {
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%d", host, port)

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	// Test connection
	_, err = client.Agent().Self()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Consul: %w", err)
	}

	log.Println("✅ Connected to Consul")

	return &ConsulClient{client: client}, nil
}

// getOutboundIP gets the preferred outbound IP of this machine
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// Register registers a service with Consul
func (c *ConsulClient) Register(cfg ServiceConfig) error {
	hostIP := getOutboundIP()

	registration := &api.AgentServiceRegistration{
		ID:      cfg.ID,
		Name:    cfg.Name,
		Port:    cfg.Port,
		Address: hostIP,
		Tags:    cfg.Tags,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", hostIP, cfg.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("✅ Registered service: %s (ID: %s) at %s:%d", cfg.Name, cfg.ID, hostIP, cfg.Port)
	return nil
}

// Deregister removes a service from Consul
func (c *ConsulClient) Deregister(serviceID string) error {
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("✅ Deregistered service: %s", serviceID)
	return nil
}

// GetService returns a healthy instance of a service
func (c *ConsulClient) GetService(serviceName string) (string, int, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get service: %w", err)
	}

	if len(services) == 0 {
		return "", 0, fmt.Errorf("no healthy instances of %s found", serviceName)
	}

	// Return first healthy instance
	service := services[0].Service
	address := service.Address
	if address == "" {
		address = "localhost"
	}

	return address, service.Port, nil
}

// GetServiceURL returns the full URL for a service
func (c *ConsulClient) GetServiceURL(serviceName string) (string, error) {
	address, port, err := c.GetService(serviceName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d", address, port), nil
}

// GetAllServices returns all registered services
func (c *ConsulClient) GetAllServices() (map[string][]string, error) {
	services, _, err := c.client.Catalog().Services(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	return services, nil
}

// Helper to get port as string
func (c *ConsulClient) GetServicePort(serviceName string) (string, error) {
	_, port, err := c.GetService(serviceName)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(port), nil
}
