package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"subscription-service/internal/dto"
)

func writeJSON(w http.ResponseWriter, logger *zap.Logger, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("failed to encode response", zap.Error(err))
	}
}

func writeError(w http.ResponseWriter, logger *zap.Logger, status int, message string) {
	writeJSON(w, logger, status, dto.ErrorResponse{Error: message})
}
