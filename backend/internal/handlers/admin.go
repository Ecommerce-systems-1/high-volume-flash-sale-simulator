package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Ecommerce-systems-1/flash-sale/internal/config"
	"github.com/Ecommerce-systems-1/flash-sale/internal/models"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

type FlashSaleInserter interface {
	InsertFlashSale(ctx context.Context, productName string, stock int, startTime, endTime time.Time) (int, error)
}

type AdminHandler struct {
	svc  *service.SaleService
	cfg  *config.Config
}

func NewAdminHandler(svc *service.SaleService, cfg *config.Config) *AdminHandler {
	return &AdminHandler{svc: svc, cfg: cfg}
}

func (h *AdminHandler) SeedHandler(dbc FlashSaleInserter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad_request"}`, http.StatusBadRequest)
			return
		}

		if req.ProductName == "" {
			req.ProductName = "Flash Sale Product"
		}
		if req.Stock < 1 || req.Stock > 10000 {
			req.Stock = h.cfg.DefaultStock
		}
		if req.DurationSeconds < 60 || req.DurationSeconds > 3600 {
			req.DurationSeconds = h.cfg.SaleDuration
		}

		startTime := time.Now()
		endTime := startTime.Add(time.Duration(req.DurationSeconds) * time.Second)

		var saleID int
		var err error
		if dbc != nil {
			saleID, err = dbc.InsertFlashSale(r.Context(), req.ProductName, req.Stock, startTime, endTime)
			if err != nil {
				log.Printf("insert flash sale error: %v", err)
				saleID = int(time.Now().UnixNano()) % 100000
			}
		} else {
			saleID = int(time.Now().UnixNano()) % 100000
		}

		if err := h.svc.InitSale(r.Context(), saleID, req.Stock, req.DurationSeconds); err != nil {
			log.Printf("init sale error: %v", err)
			http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SeedResponse{
			SaleID:    saleID,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
}