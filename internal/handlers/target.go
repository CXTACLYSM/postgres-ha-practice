package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/CXTACLYSM/postgres-ha-practice/internal/queries/slow"
	"go.uber.org/zap"
)

type TargetHandler struct {
	logger *zap.Logger
	slow   slow.Handler
}

func NewTargetHandler(logger *zap.Logger, slow slow.Handler) *TargetHandler {
	return &TargetHandler{
		logger: logger.Named("target_handler"),
		slow:   slow,
	}
}

func (h *TargetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 150*time.Millisecond)
	defer cancel()

	err := h.slow.Handle(ctx, slow.Query{})
	if err != nil {
		h.logger.Error("slow query failed",
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
