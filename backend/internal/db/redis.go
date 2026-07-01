package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	client *redis.Client
}

func NewRedis(addr string) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		PoolSize: 100,
	})
	return &RedisDB{client: client}, nil
}

func (r *RedisDB) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisDB) Close() error {
	return r.client.Close()
}

func (r *RedisDB) Client() *redis.Client {
	return r.client
}

func (r *RedisDB) StockKey(saleID int) string {
	return fmt.Sprintf("sale:%d:stock", saleID)
}

func (r *RedisDB) RequestsKey(saleID int) string {
	return fmt.Sprintf("sale:%d:requests", saleID)
}

func (r *RedisDB) SuccessKey(saleID int) string {
	return fmt.Sprintf("sale:%d:success", saleID)
}

func (r *RedisDB) ActiveKey(saleID int) string {
	return fmt.Sprintf("sale:%d:active", saleID)
}