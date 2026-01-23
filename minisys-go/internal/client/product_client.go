package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
)

type ProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetProduct fetches a product from Product Service
func (c *ProductClient) GetProduct(productID int) (*models.Product, error) {
	url := fmt.Sprintf("%s/products/%d", c.baseURL, productID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product %d not found", productID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var product models.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &product, nil
}
