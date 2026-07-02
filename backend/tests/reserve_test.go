package tests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ecommerce-systems-1/flash-sale/internal/handlers"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

func TestReserveHandlerReturns200OnSuccess(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})
	svc.InitSale(nil, 1, 10, 300)

	router := handlers.NewRouter(svc)

	body := `{"user_id":"user_1","sale_id":1}`
	req := httptest.NewRequest("POST", "/api/reserve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["order_id"] == nil {
		t.Fatal("response missing order_id")
	}
}

func TestReserveHandlerReturns409WhenSoldOut(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})
	svc.InitSale(nil, 1, 1, 300)

	// Consume the only item
	svc.AttemptReservation(nil, 1)

	router := handlers.NewRouter(svc)

	body := `{"user_id":"user_1","sale_id":1}`
	req := httptest.NewRequest("POST", "/api/reserve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 409 {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestReserveHandlerReturns410WhenExpired(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})
	// Don't initialize the sale at all - no active key exists
	// Just try to reserve for a non-existent sale

	router := handlers.NewRouter(svc)

	body := `{"user_id":"user_1","sale_id":999}`
	req := httptest.NewRequest("POST", "/api/reserve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 410 {
		t.Fatalf("expected 410, got %d", w.Code)
	}
}

func TestStatsHandler(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})
	svc.InitSale(nil, 1, 50, 300)

	router := handlers.NewRouter(svc)

	req := httptest.NewRequest("GET", "/api/stats?sale_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["stock_remaining"] == nil {
		t.Fatal("response missing stock_remaining")
	}
}

func TestHealthEndpoint(t *testing.T) {
	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})

	router := handlers.NewRouter(svc)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", resp["status"])
	}
}
