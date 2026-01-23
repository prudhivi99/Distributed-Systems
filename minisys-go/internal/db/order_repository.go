package db

import (
	"database/sql"
	"fmt"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(database *PostgresDB) *OrderRepository {
	return &OrderRepository{db: database.Conn}
}

// Create inserts a new order with items
func (r *OrderRepository) Create(order *models.Order) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (customer_name, total_amount, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err = tx.QueryRow(orderQuery, order.CustomerName, order.TotalAmount, order.Status).
		Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, product_name, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		err = tx.QueryRow(itemQuery,
			order.ID,
			order.Items[i].ProductID,
			order.Items[i].ProductName,
			order.Items[i].Quantity,
			order.Items[i].Price,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAll returns all orders
func (r *OrderRepository) GetAll() ([]models.Order, error) {
	query := `SELECT id, customer_name, total_amount, status, created_at FROM orders ORDER BY id DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.CustomerName, &o.TotalAmount, &o.Status, &o.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, nil
}

// GetByID returns a single order with items
func (r *OrderRepository) GetByID(id int) (*models.Order, error) {
	// Get order
	orderQuery := `SELECT id, customer_name, total_amount, status, created_at FROM orders WHERE id = $1`

	var order models.Order
	err := r.db.QueryRow(orderQuery, id).
		Scan(&order.ID, &order.CustomerName, &order.TotalAmount, &order.Status, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	itemsQuery := `SELECT id, order_id, product_id, product_name, quantity, price FROM order_items WHERE order_id = $1`

	rows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// UpdateStatus updates order status
func (r *OrderRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}
