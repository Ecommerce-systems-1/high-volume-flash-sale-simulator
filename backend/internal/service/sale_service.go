package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrSoldOut     = errors.New("sold_out")
	ErrSaleExpired = errors.New("sale_expired")
)

type OrderCreator interface {
	CreateOrder(ctx context.Context, saleID int, userID string) (string, error)
}

// SaleStore provides all storage operations using sync.Map (no external Redis/PostgreSQL needed)
type SaleStore struct {
	mu         sync.Mutex
	stock      map[int]int64
	requests   map[int]int64
	successes  map[int]int64
	actives    map[int]time.Time
	durations  map[int]time.Duration
	orderCount map[int]int
}

func NewSaleStore() *SaleStore {
	return &SaleStore{
		stock:      make(map[int]int64),
		requests:   make(map[int]int64),
		successes:  make(map[int]int64),
		actives:    make(map[int]time.Time),
		durations:  make(map[int]time.Duration),
		orderCount: make(map[int]int),
	}
}

func (s *SaleStore) InitSale(saleID int, stock int64, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stock[saleID] = stock
	s.requests[saleID] = 0
	s.successes[saleID] = 0
	s.actives[saleID] = time.Now().Add(duration)
	s.durations[saleID] = duration
}

func (s *SaleStore) IsActive(saleID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, ok := s.actives[saleID]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}

func (s *SaleStore) AttemptReservation(saleID int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check active
	expiry, ok := s.actives[saleID]
	if !ok || time.Now().After(expiry) {
		return 0, ErrSaleExpired
	}

	// Atomic DECR equivalent
	stock, ok := s.stock[saleID]
	if !ok {
		return 0, ErrSaleExpired
	}
	if stock <= 0 {
		return 0, ErrSoldOut
	}

	s.stock[saleID] = stock - 1
	s.successes[saleID]++
	s.requests[saleID]++
	return stock - 1, nil
}

func (s *SaleStore) IncrementRequests(saleID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests[saleID]++
}

func (s *SaleStore) GetStock(saleID int) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stock[saleID]
}

func (s *SaleStore) GetRequests(saleID int) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.requests[saleID]
}

func (s *SaleStore) GetSuccesses(saleID int) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.successes[saleID]
}

func (s *SaleStore) GetTTL(saleID int) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, ok := s.actives[saleID]
	if !ok {
		return 0
	}
	remaining := time.Until(expiry).Seconds()
	if remaining < 0 {
		return 0
	}
	return int64(remaining)
}

func (s *SaleStore) IncrementOrderCount(saleID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orderCount[saleID]++
}

type SaleService struct {
	store *SaleStore
	dbc   OrderCreator
}

func NewSaleService(store *SaleStore, dbc OrderCreator) *SaleService {
	return &SaleService{store: store, dbc: dbc}
}

func (s *SaleService) AttemptReservation(ctx context.Context, saleID int) (int64, error) {
	return s.store.AttemptReservation(saleID)
}

func (s *SaleService) CreateOrderRecord(ctx context.Context, saleID int, userID string) (string, error) {
	s.store.IncrementOrderCount(saleID)
	return s.dbc.CreateOrder(ctx, saleID, userID)
}

func (s *SaleService) GetReserveStats(ctx context.Context, saleID int) (int64, int64, int64, error) {
	stock := s.store.GetStock(saleID)
	successes := s.store.GetSuccesses(saleID)
	requests := s.store.GetRequests(saleID)
	rejectedSoldOut := requests - successes
	if rejectedSoldOut < 0 {
		rejectedSoldOut = 0
	}
	return stock, successes, rejectedSoldOut, nil
}

func (s *SaleService) GetTTL(ctx context.Context, saleID int) (int64, error) {
	return s.store.GetTTL(saleID), nil
}

func (s *SaleService) InitSale(ctx context.Context, saleID int, stock int, ttlSeconds int) error {
	s.store.InitSale(saleID, int64(stock), time.Duration(ttlSeconds)*time.Second)
	return nil
}

func (s *SaleService) IncrementRequests(ctx context.Context, saleID int) {
	s.store.IncrementRequests(saleID)
}
