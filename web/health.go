package web

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Health struct {
	UptimeSeconds uint64 `json:"uptimeSeconds"`
	APIRequests   uint64 `json:"apiRequests"`
}

func handleHealthRequest(writer http.ResponseWriter, request *http.Request) {
	b, err := json.Marshal(
		Health{
			APIRequests:   apiRequests,
			UptimeSeconds: uint64(time.Since(startupTime).Seconds()),
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
