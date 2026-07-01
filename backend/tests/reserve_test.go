package tests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ecommerce-systems-1/flash-sale/internal/handlers"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestReserveHandlerReturns200OnSuccess(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mr.Set("sale:1:stock", "10")
	mr.Set("sale:1:active", "1")
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := service.NewSaleService(rdb, &mockDB{})
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
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mr.Set("sale:1:stock", "0")
	mr.Set("sale:1:active", "1")
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := service.NewSaleService(rdb, &mockDB{})
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
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mr.Set("sale:1:stock", "10")
	// Don't set active key - sale has expired
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := service.NewSaleService(rdb, &mockDB{})
	router := handlers.NewRouter(svc)

	body := `{"user_id":"user_1","sale_id":1}`
	req := httptest.NewRequest("POST", "/api/reserve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 410 {
		t.Fatalf("expected 410, got %d", w.Code)
	}
}

func TestStatsHandler(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	mr.Set("sale:1:stock", "50")
	mr.Set("sale:1:requests", "30")
	mr.Set("sale:1:success", "20")
	mr.Set("sale:1:active", "1")
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := service.NewSaleService(rdb, &mockDB{})
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
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := service.NewSaleService(rdb, &mockDB{})
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
