package metrics

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HttpRequestsTotal         *prometheus.CounterVec
	PgxPoolReadAcquiredConns  prometheus.Gauge
	PgxPoolWriteAcquiredConns prometheus.Gauge
	HttpRequestDuration       *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by status",
		},
		[]string{"status"},
	)
	pgxpoolReadAcquiredConns := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "pgxpool_read_acquired_conns",
			Help: "Number of currently acquired connections in the pool",
		},
	)
	pgxpoolWriteAcquiredConns := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "pgxpool_write_acquired_conns",
			Help: "Number of currently acquired connections in the pool",
		},
	)
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(pgxpoolReadAcquiredConns)
	prometheus.MustRegister(pgxpoolWriteAcquiredConns)
	prometheus.MustRegister(httpRequestDuration)

	return &Metrics{
		HttpRequestsTotal:         httpRequestsTotal,
		PgxPoolReadAcquiredConns:  pgxpoolReadAcquiredConns,
		PgxPoolWriteAcquiredConns: pgxpoolWriteAcquiredConns,
		HttpRequestDuration:       httpRequestDuration,
	}
}

func StartPoolMetricsCollector(
	m *Metrics,
	readPool *pgxpool.Pool,
	writePool *pgxpool.Pool,
) {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			readAcquired := readPool.Stat().AcquiredConns()
			writeAcquired := writePool.Stat().AcquiredConns()

			m.PgxPoolReadAcquiredConns.Set(float64(readAcquired))
			m.PgxPoolWriteAcquiredConns.Set(float64(writeAcquired))
		}
	}()
}
