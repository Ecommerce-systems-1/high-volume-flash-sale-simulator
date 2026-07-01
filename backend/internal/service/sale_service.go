package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSoldOut     = errors.New("sold_out")
	ErrSaleExpired = errors.New("sale_expired")
)

type OrderCreator interface {
	CreateOrder(ctx context.Context, saleID int, userID string) (string, error)
}

type SaleService struct {
	rdb *redis.Client
	dbc OrderCreator
}

func NewSaleService(rdb *redis.Client, dbc OrderCreator) *SaleService {
	return &SaleService{rdb: rdb, dbc: dbc}
}

func (s *SaleService) AttemptReservation(ctx context.Context, saleID int) (int64, error) {
	activeKey := fmt.Sprintf("sale:%d:active", saleID)
	stockKey := fmt.Sprintf("sale:%d:stock", saleID)

	exists, err := s.rdb.Exists(ctx, activeKey).Result()
	if err != nil {
		return 0, err
	}
	if exists == 0 {
		return 0, ErrSaleExpired
	}

	remaining, err := s.rdb.Decr(ctx, stockKey).Result()
	if err != nil {
		return 0, err
	}
	if remaining < 0 {
		s.rdb.Incr(ctx, stockKey)
		return 0, ErrSoldOut
	}

	s.rdb.Incr(ctx, fmt.Sprintf("sale:%d:success", saleID))
	return remaining, nil
}

func (s *SaleService) CreateOrderRecord(ctx context.Context, saleID int, userID string) (string, error) {
	return s.dbc.CreateOrder(ctx, saleID, userID)
}

func (s *SaleService) GetReserveStats(ctx context.Context, saleID int) (int64, int64, int64, error) {
	stockKey := fmt.Sprintf("sale:%d:stock", saleID)
	requestsKey := fmt.Sprintf("sale:%d:requests", saleID)
	successKey := fmt.Sprintf("sale:%d:success", saleID)

	stock, err := s.rdb.Get(ctx, stockKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	requests, err := s.rdb.Get(ctx, requestsKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	success, err := s.rdb.Get(ctx, successKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	rejectedSoldOut := requests - success
	if rejectedSoldOut < 0 {
		rejectedSoldOut = 0
	}

	return stock, success, rejectedSoldOut, nil
}

func (s *SaleService) GetTTL(ctx context.Context, saleID int) (int64, error) {
	activeKey := fmt.Sprintf("sale:%d:active", saleID)
	ttl, err := s.rdb.TTL(ctx, activeKey).Result()
	if err != nil {
		return 0, err
	}
	return int64(ttl.Seconds()), nil
}

func (s *SaleService) InitSale(ctx context.Context, saleID int, stock int, ttlSeconds int) error {
	stockKey := fmt.Sprintf("sale:%d:stock", saleID)
	requestsKey := fmt.Sprintf("sale:%d:requests", saleID)
	successKey := fmt.Sprintf("sale:%d:success", saleID)
	activeKey := fmt.Sprintf("sale:%d:active", saleID)

	if err := s.rdb.Set(ctx, stockKey, stock, 0).Err(); err != nil {
		return err
	}
	if err := s.rdb.Set(ctx, requestsKey, 0, 0).Err(); err != nil {
		return err
	}
	if err := s.rdb.Set(ctx, successKey, 0, 0).Err(); err != nil {
		return err
	}
	if err := s.rdb.Set(ctx, activeKey, "1", 0).Err(); err != nil {
		return err
	}
	if err := s.rdb.Expire(ctx, activeKey, time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
		return err
	}
	return nil
}

func (s *SaleService) IncrementRequests(ctx context.Context, saleID int) {
	s.rdb.Incr(ctx, fmt.Sprintf("sale:%d:requests", saleID))
}
