package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ecommerce-systems-1/flash-sale/internal/config"
	"github.com/Ecommerce-systems-1/flash-sale/internal/db"
	"github.com/Ecommerce-systems-1/flash-sale/internal/handlers"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

func main() {
	cfg := config.Load()

	// Connect to Redis
	rdb, err := db.NewRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// Connect to PostgreSQL
	pdb, err := db.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Printf("warning: failed to connect to postgres: %v (running without DB)", err)
	}
	defer func() {
		if pdb != nil {
			pdb.Close()
		}
	}()

	// Create sale service
	svc := service.NewSaleService(rdb.Client(), pdb)

	// Seed default sale
	ctx := context.Background()
	saleID := 1
	if err := svc.InitSale(ctx, saleID, cfg.DefaultStock, cfg.SaleDuration); err != nil {
		log.Printf("warning: failed to seed default sale: %v", err)
	} else {
		log.Printf("seeded default flash sale: id=%d, stock=%d, duration=%ds", saleID, cfg.DefaultStock, cfg.SaleDuration)
	}

	// Create router
	router := handlers.NewRouter(svc)

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
