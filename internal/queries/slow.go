package queries

import (
	"context"
	"fmt"

	"github.com/CXTACLYSM/postgres-ha-practice/internal/queries/slow"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlowQueryHandler struct {
	pool *pgxpool.Pool
}

func NewSlowQueryHandler(pool *pgxpool.Pool) *SlowQueryHandler {
	return &SlowQueryHandler{
		pool: pool,
	}
}

func (h *SlowQueryHandler) Handle(ctx context.Context, query slow.Query) error {
	_, err := h.pool.Exec(ctx, "SELECT pg_sleep(0.100)")
	if err != nil {
		return fmt.Errorf("error executing slow query: %w", err)
	}

	return nil
}
