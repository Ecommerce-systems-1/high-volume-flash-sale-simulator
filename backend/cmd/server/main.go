package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Ecommerce-systems-1/flash-sale/internal/config"
	"github.com/Ecommerce-systems-1/flash-sale/internal/handlers"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

type mockDB struct{}

func (m *mockDB) CreateOrder(ctx context.Context, saleID int, userID string) (string, error) {
	return "mock-order-id", nil
}

func main() {
	cfg := config.Load()

	store := service.NewSaleStore()
	svc := service.NewSaleService(store, &mockDB{})

	// Seed default sale
	ctx := context.Background()
	saleID := 1
	if err := svc.InitSale(ctx, saleID, cfg.DefaultStock, cfg.SaleDuration); err != nil {
		log.Printf("warning: failed to seed default sale: %v", err)
	} else {
		log.Printf("seeded default flash sale: id=%d, stock=%d, duration=%ds", saleID, cfg.DefaultStock, cfg.SaleDuration)
	}

	// Create router with static file serving
	router := handlers.NewRouter(svc)

	// Try to serve static frontend files
	staticDir := filepath.Join("..", "frontend", "out")
	if _, err := os.Stat(staticDir); err == nil {
		fileServer := http.FileServer(http.Dir(staticDir))
		router = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For API and health routes, use the normal handler
			if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
				handlers.NewRouter(svc).ServeHTTP(w, r)
				return
			}
			if r.URL.Path == "/health" {
				handlers.NewRouter(svc).ServeHTTP(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("flash sale server listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
