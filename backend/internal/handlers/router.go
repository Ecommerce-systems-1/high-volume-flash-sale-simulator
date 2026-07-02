package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/Ecommerce-systems-1/flash-sale/internal/service"
)

var StaticDir string

func NewRouter(svc *service.SaleService) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/reserve", ReserveHandler(svc))
	mux.HandleFunc("/api/stats", StatsHandler(svc))
	mux.HandleFunc("/health", HealthHandler)

	// Try multiple possible static dir locations
	candidates := []string{
		StaticDir,
		filepath.Join("..", "frontend", "out"),
		"/frontend/out",
	}
	for _, d := range candidates {
		if d != "" {
			if fi, err := os.Stat(d); err == nil && fi.IsDir() {
				fileServer := http.FileServer(http.Dir(d))
				mux.Handle("/", fileServer)
				break
			}
		}
	}

	return mux
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
