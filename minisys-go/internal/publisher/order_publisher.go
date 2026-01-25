package publisher

import (
	"encoding/json"
	"fmt"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/messaging"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
)

const OrderCreatedQueue = "order.created"

type OrderPublisher struct {
	mq *messaging.RabbitMQ
}

func NewOrderPublisher(mq *messaging.RabbitMQ) (*OrderPublisher, error) {
	// Declare the queue
	if err := mq.DeclareQueue(OrderCreatedQueue); err != nil {
		return nil, err
	}

	return &OrderPublisher{mq: mq}, nil
}

// PublishOrderCreated publishes an order.created event
func (p *OrderPublisher) PublishOrderCreated(order *models.Order) error {
	event := models.OrderCreatedEvent{
		OrderID:      order.ID,
		CustomerName: order.CustomerName,
		TotalAmount:  order.TotalAmount,
	}

	for _, item := range order.Items {
		event.Items = append(event.Items, models.OrderItemEvent{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.mq.Publish(OrderCreatedQueue, data)
}
