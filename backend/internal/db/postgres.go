package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) CreateOrder(ctx context.Context, saleID int, userID string) (string, error) {
	var id string
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO orders (sale_id, user_id, status) VALUES ($1, $2, 'RESERVED') RETURNING id`,
		saleID, userID,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create order: %w", err)
	}
	return id, nil
}

func (p *PostgresDB) InsertFlashSale(ctx context.Context, productName string, stock int, startTime, endTime time.Time) (int, error) {
	var id int
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO flash_sales (product_id, product_name, initial_stock, start_time, end_time) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		fmt.Sprintf("prod_%d", time.Now().UnixNano()), productName, stock, startTime, endTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert flash sale: %w", err)
	}
	return id, nil
}

func (p *PostgresDB) GetFlashSale(ctx context.Context, saleID int) (*struct {
	ID           int
	ProductName  string
	InitialStock int
	StartTime    time.Time
	EndTime      time.Time
}, error) {
	var result struct {
		ID           int
		ProductName  string
		InitialStock int
		StartTime    time.Time
		EndTime      time.Time
	}
	err := p.db.QueryRowContext(ctx,
		`SELECT id, product_name, initial_stock, start_time, end_time FROM flash_sales WHERE id = $1`,
		saleID,
	).Scan(&result.ID, &result.ProductName, &result.InitialStock, &result.StartTime, &result.EndTime)
	if err != nil {
		return nil, fmt.Errorf("get flash sale: %w", err)
	}
	return &result, nil
}

func (p *PostgresDB) GetOrderCount(ctx context.Context, saleID int) (int, error) {
	var count int
	err := p.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders WHERE sale_id = $1 AND status = 'RESERVED'`,
		saleID,
	).Scan(&count)
	return count, err
}