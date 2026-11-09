package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/satellite-communications/gateway/internal/model"
)

func NewHandler(provider SatelliteProvider) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	})

	mux.HandleFunc("GET /api/satellites", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			map[string]any{
				"satellites": provider.GetSnapshot(),
			},
		); err != nil {
			return
		}
	})

	return mux
}

type SatelliteProvider interface {
	GetSnapshot() []model.SatelliteSnapshot
}
