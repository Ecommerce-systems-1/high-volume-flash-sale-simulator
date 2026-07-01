package models

import "time"

type FlashSale struct {
	ID           int       `json:"id"`
	ProductID    string    `json:"product_id"`
	ProductName  string    `json:"product_name"`
	InitialStock int       `json:"initial_stock"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	CreatedAt    time.Time `json:"created_at"`
}

type Order struct {
	ID         string    `json:"id"`
	SaleID     int       `json:"sale_id"`
	UserID     string    `json:"user_id"`
	ReservedAt time.Time `json:"reserved_at"`
	Status     string    `json:"status"`
}

type StatsSnapshot struct {
	ID               int       `json:"id"`
	SaleID           int       `json:"sale_id"`
	SnapshotTime     time.Time `json:"snapshot_time"`
	TotalRequests    int       `json:"total_requests"`
	Successful       int       `json:"successful"`
	RejectedSoldOut  int       `json:"rejected_sold_out"`
	RejectedExpired  int       `json:"rejected_expired"`
	RPS              float64   `json:"rps"`
}

type ReserveRequest struct {
	UserID string `json:"user_id"`
	SaleID int    `json:"sale_id"`
}

type ReserveResponse struct {
	OrderID        string `json:"order_id"`
	StockRemaining int64  `json:"stock_remaining"`
}

type StatsResponse struct {
	SaleID          int     `json:"sale_id"`
	ProductName     string  `json:"product_name"`
	InitialStock    int     `json:"initial_stock"`
	StockRemaining  int64   `json:"stock_remaining"`
	TotalRequests   int64   `json:"total_requests"`
	Successful      int64   `json:"successful"`
	RejectedSoldOut int64   `json:"rejected_sold_out"`
	RejectedExpired int64   `json:"rejected_expired"`
	RPS             float64 `json:"rps"`
	SaleActive      bool    `json:"sale_active"`
	SecondsRemaining int64  `json:"seconds_remaining"`
}

type SeedRequest struct {
	ProductName     string `json:"product_name"`
	Stock           int    `json:"stock"`
	DurationSeconds int    `json:"duration_seconds"`
}

type SeedResponse struct {
	SaleID    int       `json:"sale_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}