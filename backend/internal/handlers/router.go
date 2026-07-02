package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

func NewRouter(svc *service.SaleService) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/reserve", ReserveHandler(svc))
	mux.HandleFunc("/api/stats", StatsHandler(svc))
	mux.HandleFunc("/health", HealthHandler)

	// Serve static frontend
	staticDir := filepath.Join("..", "frontend", "out")
	if _, err := os.Stat(staticDir); err == nil {
		fileServer := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", fileServer)
	}

	return mux
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
