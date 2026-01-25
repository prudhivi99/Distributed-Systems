package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/cache"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type InventoryConsumer struct {
	repo  *db.ProductRepository
	cache *cache.RedisCache
}

func NewInventoryConsumer(repo *db.ProductRepository, cache *cache.RedisCache) *InventoryConsumer {
	return &InventoryConsumer{
		repo:  repo,
		cache: cache,
	}
}

// ProcessOrderCreated handles order.created events
func (c *InventoryConsumer) ProcessOrderCreated(messages <-chan amqp.Delivery) {
	for msg := range messages {
		log.Printf("📥 Received order.created event")

		var event models.OrderCreatedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("❌ Failed to parse event: %v", err)
			msg.Nack(false, false)
			continue
		}

		log.Printf("📦 Processing Order #%d for %s", event.OrderID, event.CustomerName)

		success := true
		ctx := context.Background()

		for _, item := range event.Items {
			err := c.repo.UpdateQuantity(item.ProductID, -item.Quantity)
			if err != nil {
				log.Printf("❌ Failed to update inventory for product %d: %v", item.ProductID, err)
				success = false
			} else {
				log.Printf("✅ Reduced inventory: Product %d by %d", item.ProductID, item.Quantity)

				// Invalidate cache
				productKey := fmt.Sprintf("product:%d", item.ProductID)
				c.cache.Delete(ctx, productKey)
				c.cache.Delete(ctx, "products:all")
				log.Printf("🗑️ Cache invalidated: %s", productKey)
			}
		}

		if success {
			msg.Ack(false)
			log.Printf("✅ Order #%d processed successfully", event.OrderID)
		} else {
			msg.Nack(false, true)
			log.Printf("⚠️ Order #%d partially failed, requeued", event.OrderID)
		}
	}
}
