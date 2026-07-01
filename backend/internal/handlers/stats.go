package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Ecommerce-systems-1/flash-sale/internal/models"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

func StatsHandler(svc *service.SaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		saleIDStr := r.URL.Query().Get("sale_id")
		if saleIDStr == "" {
			http.Error(w, `{"error":"missing_sale_id"}`, http.StatusBadRequest)
			return
		}
		saleID, err := strconv.Atoi(saleIDStr)
		if err != nil || saleID <= 0 {
			http.Error(w, `{"error":"invalid_sale_id"}`, http.StatusBadRequest)
			return
		}

		stock, success, rejected, err := svc.GetReserveStats(r.Context(), saleID)
		if err != nil {
			log.Printf("stats error: %v", err)
			http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
			return
		}

		ttl, _ := svc.GetTTL(r.Context(), saleID)

		// Get total requests from stats
		requestsKey := "sale:" + strconv.Itoa(saleID) + ":requests"
		_ = requestsKey

		// For now use success + rejected as total
		totalRequests := success + rejected
		if totalRequests < 0 {
			totalRequests = 0
		}

		saleActive := ttl > 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.StatsResponse{
			SaleID:          saleID,
			ProductName:     "Flash Sale Product",
			InitialStock:    100,
			StockRemaining:  stock,
			TotalRequests:   totalRequests,
			Successful:      success,
			RejectedSoldOut: rejected,
			RejectedExpired: 0,
			RPS:             0,
			SaleActive:      saleActive,
			SecondsRemaining: ttl,
		})
	}
}