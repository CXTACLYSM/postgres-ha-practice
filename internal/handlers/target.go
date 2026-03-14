package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/CXTACLYSM/postgres-ha-practice/internal/queries/slow"
)

type TargetHandler struct {
	Slow slow.Handler
}

func NewTargetHandler(slow slow.Handler) *TargetHandler {
	return &TargetHandler{
		Slow: slow,
	}
}

func (h *TargetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 150*time.Millisecond)
	defer cancel()

	err := h.Slow.Handle(ctx, slow.Query{})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
