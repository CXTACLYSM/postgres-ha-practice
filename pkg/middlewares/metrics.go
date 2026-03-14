package middlewares

import (
	"net/http"
	"time"

	"github.com/CXTACLYSM/postgres-ha-practice/pkg/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

type Metrics struct {
	prometheus   *metrics.Metrics
	readPgxPool  *pgxpool.Pool
	writePgxPool *pgxpool.Pool
}

func NewMetricsMiddleware(prometheus *metrics.Metrics, readPgxPool *pgxpool.Pool, writePgxPool *pgxpool.Pool) *Metrics {
	return &Metrics{
		prometheus:   prometheus,
		readPgxPool:  readPgxPool,
		writePgxPool: writePgxPool,
	}
}

func (m *Metrics) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start).Seconds()

		status := "success"
		if rec.status == http.StatusServiceUnavailable || rec.status == http.StatusGatewayTimeout {
			status = "timeout"
		}

		m.prometheus.HttpRequestsTotal.WithLabelValues(status).Inc()
		m.prometheus.HttpRequestDuration.WithLabelValues(status).Observe(elapsed)
	}
	return http.HandlerFunc(fn)
}
