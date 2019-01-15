package web

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
	"raspi_readtemp/logging"
	"time"
)

var apiRequests uint64
var startupTime time.Time
var logger = logging.New("raspi_temperature_service_web", false)

func SetupChi(chiRouter *chi.Mux) error {
	startupTime = time.Now()
	chiRouter.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// very basic metric
			apiRequests++
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	chiRouter.Get("/health", handleHealthRequest)
	return nil
}

func handleHealthRequest(writer http.ResponseWriter, request *http.Request) {
	b, err := json.Marshal(
		Health{
			APIRequests: apiRequests,
			UptimeSeconds: uint64(time.Now().Sub(startupTime).Seconds()),
		},
	)

	if err != nil {
		logger.Error("Cannot marshall data", zap.String("err", err.Error()))
		http.Error(writer, err.Error(), 500)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.Write(b)
}

// ==== MODEL ===
type Health struct {
	UptimeSeconds uint64 `json:"uptimeSeconds"`
	APIRequests   uint64 `json:"apiRequests"`
}
