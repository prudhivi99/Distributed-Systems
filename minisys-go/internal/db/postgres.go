package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	Conn *sql.DB
}

func NewPostgresDB(host string, port int, user, password, dbname string) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return &PostgresDB{Conn: conn}, nil
}

func (db *PostgresDB) Close() error {
	return db.Conn.Close()
}
