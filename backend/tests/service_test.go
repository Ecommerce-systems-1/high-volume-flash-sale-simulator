package tests

import (
	"context"
	"testing"

	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

type mockDB struct{}

func (m *mockDB) CreateOrder(ctx context.Context, saleID int, userID string) (string, error) {
	return "mock-order-id", nil
}

func TestAtomicReservation(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})

	ctx := context.Background()
	saleID := 1

	// Initialize sale with 10 items
	if err := svc.InitSale(ctx, saleID, 10, 300); err != nil {
		t.Fatalf("InitSale failed: %v", err)
	}

	// Reserve all 10 items
	for i := 0; i < 10; i++ {
		remaining, err := svc.AttemptReservation(ctx, saleID)
		if err != nil {
			t.Fatalf("AttemptReservation failed at iteration %d: %v", i, err)
		}
		if remaining < 0 {
			t.Fatalf("unexpected negative remaining at iteration %d: %d", i, remaining)
		}
	}

	// 11th attempt should fail with sold out
	_, err := svc.AttemptReservation(ctx, saleID)
	if err != service.ErrSoldOut {
		t.Fatalf("expected ErrSoldOut, got %v", err)
	}

	// Verify stats
	stock, success, _, err := svc.GetReserveStats(ctx, saleID)
	if err != nil {
		t.Fatalf("GetReserveStats failed: %v", err)
	}
	if stock != 0 {
		t.Fatalf("expected stock 0, got %d", stock)
	}
	if success != 10 {
		t.Fatalf("expected 10 successful, got %d", success)
	}
}

func TestReservationRejectsWhenStockZero(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})

	ctx := context.Background()
	saleID := 2

	// Initialize with 1 item
	if err := svc.InitSale(ctx, saleID, 1, 300); err != nil {
		t.Fatalf("InitSale failed: %v", err)
	}

	// Consume the only item
	_, err := svc.AttemptReservation(ctx, saleID)
	if err != nil {
		t.Fatalf("first reservation should succeed: %v", err)
	}

	// Second attempt should fail
	_, err = svc.AttemptReservation(ctx, saleID)
	if err != service.ErrSoldOut {
		t.Fatalf("expected ErrSoldOut on second attempt, got %v", err)
	}
}
