package models

// OrderCreatedEvent is published when a new order is created
type OrderCreatedEvent struct {
	OrderID      int              `json:"order_id"`
	CustomerName string           `json:"customer_name"`
	TotalAmount  float64          `json:"total_amount"`
	Items        []OrderItemEvent `json:"items"`
}

type OrderItemEvent struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// InventoryUpdateEvent is for updating product inventory
type InventoryUpdateEvent struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"` // negative = reduce, positive = add
}
