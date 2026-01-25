package consumer

import (
	"encoding/json"
	"log"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/db"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type InventoryConsumer struct {
	repo *db.ProductRepository
}

func NewInventoryConsumer(repo *db.ProductRepository) *InventoryConsumer {
	return &InventoryConsumer{repo: repo}
}

// ProcessOrderCreated handles order.created events
func (c *InventoryConsumer) ProcessOrderCreated(messages <-chan amqp.Delivery) {
	for msg := range messages {
		log.Printf("📥 Received order.created event")

		var event models.OrderCreatedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("❌ Failed to parse event: %v", err)
			msg.Nack(false, false) // Don't requeue bad messages
			continue
		}

		log.Printf("📦 Processing Order #%d for %s", event.OrderID, event.CustomerName)

		// Update inventory for each item
		success := true
		for _, item := range event.Items {
			// Reduce inventory (negative quantity)
			err := c.repo.UpdateQuantity(item.ProductID, -item.Quantity)
			if err != nil {
				log.Printf("❌ Failed to update inventory for product %d: %v", item.ProductID, err)
				success = false
			} else {
				log.Printf("✅ Reduced inventory: Product %d by %d", item.ProductID, item.Quantity)
			}
		}

		if success {
			msg.Ack(false) // Acknowledge message
			log.Printf("✅ Order #%d processed successfully", event.OrderID)
		} else {
			msg.Nack(false, true) // Requeue for retry
			log.Printf("⚠️ Order #%d partially failed, requeued", event.OrderID)
		}
	}
}
