package models

import "time"

type Product struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProductRequest struct {
	Name     string  `json:"name" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Quantity int     `json:"Quantity"`
}
