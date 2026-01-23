package models

import "time"

type Order struct {
	ID           int         `json:"id"`
	CustomerName string      `json:"customer_name"`
	TotalAmount  float64     `json:"total_amount"`
	Status       string      `json:"status"`
	Items        []OrderItem `json:"items"`
	CreatedAt    time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID          int     `json:"id"`
	OrderID     int     `json:"order_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name" binding:"required"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required"`
}

type CreateOrderItemRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required"`
}
