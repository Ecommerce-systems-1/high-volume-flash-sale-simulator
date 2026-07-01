package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Ecommerce-systems-1/flash-sale/internal/models"
	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

func ReserveHandler(svc *service.SaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ReserveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad_request"}`, http.StatusBadRequest)
			return
		}

		if req.UserID == "" || len(req.UserID) > 50 {
			http.Error(w, `{"error":"invalid_user_id"}`, http.StatusBadRequest)
			return
		}
		if req.SaleID <= 0 {
			http.Error(w, `{"error":"invalid_sale_id"}`, http.StatusBadRequest)
			return
		}

		svc.IncrementRequests(r.Context(), req.SaleID)

		remaining, err := svc.AttemptReservation(r.Context(), req.SaleID)
		if err == service.ErrSoldOut {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "sold_out", "message": "All units have been reserved"})
			return
		}
		if err == service.ErrSaleExpired {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone)
			json.NewEncoder(w).Encode(map[string]string{"error": "sale_expired", "message": "Flash sale has ended"})
			return
		}
		if err != nil {
			log.Printf("reservation error: %v", err)
			http.Error(w, `{"error":"internal"}`, http.StatusServiceUnavailable)
			return
		}

		orderID, err := svc.CreateOrderRecord(r.Context(), req.SaleID, req.UserID)
		if err != nil {
			log.Printf("order creation error: %v", err)
			orderID = fmt.Sprintf("pending-%d-%s", req.SaleID, req.UserID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ReserveResponse{
			OrderID:        orderID,
			StockRemaining: remaining,
		})
	}
}