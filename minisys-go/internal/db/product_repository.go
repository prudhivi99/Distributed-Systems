package db

import (
	"database/sql"
	"fmt"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(database *PostgresDB) *ProductRepository {
	return &ProductRepository{db: database.Conn}
}

// GetAll returns all products
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	query := "SELECT id, name, price, quantity, created_at FROM products ORDER BY id"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}

// GetByID returns a single product
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	query := "SELECT id, name, price, quantity, created_at FROM products WHERE id = $1"

	var p models.Product
	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &p, nil
}

// Create inserts a new product
func (r *ProductRepository) Create(req models.CreateProductRequest) (*models.Product, error) {
	query := `
		INSERT INTO products (name, price, quantity)
		VALUES ($1, $2, $3)
		RETURNING id, name, price, quantity, created_at
	`

	var p models.Product
	err := r.db.QueryRow(query, req.Name, req.Price, req.Quantity).
		Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &p, nil
}

// Delete removes a product
func (r *ProductRepository) Delete(id int) error {
	query := "DELETE FROM products WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}
