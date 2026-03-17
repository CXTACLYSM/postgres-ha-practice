package di

import (
	"fmt"
	"net/http"

	"github.com/CXTACLYSM/postgres-ha-practice/configs"
	"github.com/CXTACLYSM/postgres-ha-practice/internal/handlers"
	"github.com/CXTACLYSM/postgres-ha-practice/internal/queries"
	"github.com/CXTACLYSM/postgres-ha-practice/internal/queries/slow"
	"github.com/CXTACLYSM/postgres-ha-practice/pkg/metrics"
	"github.com/CXTACLYSM/postgres-ha-practice/pkg/middlewares"
	"github.com/CXTACLYSM/postgres-ha-practice/pkg/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Infrastructure struct {
	PgConnector *postgres.Connector
	Logger      *zap.Logger
	Metrics     *metrics.Metrics
}

type Queries struct {
	Slow slow.Handler
}

type Middlewares struct {
	Metrics *middlewares.Metrics
}

type Handlers struct {
	Target  *handlers.TargetHandler
	Metrics http.Handler
}

type Container struct {
	Infrastructure *Infrastructure
	Middlewares    *Middlewares
	Handlers       *Handlers
	Queries        *Queries
}

func (c *Container) Init(cfg *configs.Config) error {
	if err := c.initInfrastructure(cfg); err != nil {
		return fmt.Errorf("error initializing container infrastructure: %w", err)
	}
	if err := c.initQueries(); err != nil {
		return fmt.Errorf("error initializing container queries: %w", err)
	}
	if err := c.initMiddlewares(); err != nil {
		return fmt.Errorf("error initializing container middlewares: %w", err)
	}
	if err := c.initHandlers(); err != nil {
		return fmt.Errorf("error initializing container handlers: %w", err)
	}
	return nil
}

func (c *Container) initInfrastructure(cfg *configs.Config) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("error creating zap logger: %w", err)
	}

	readDSN, err := cfg.PostgresCluster.DSN(postgres.ReadOperation)
	if err != nil {
		return fmt.Errorf("error getting read dsn error: %w", err)
	}
	writeDSN, err := cfg.PostgresCluster.DSN(postgres.WriteOperation)
	if err != nil {
		return fmt.Errorf("error getting write dsn error: %w", err)
	}

	pgConnector, err := postgres.NewConnector(readDSN, writeDSN)
	if err != nil {
		return fmt.Errorf("error creating cluster connector: %w", err)
	}

	promMetrics := metrics.NewMetrics()
	metrics.StartPoolMetricsCollector(promMetrics, pgConnector.ReadPool, pgConnector.WritePool)

	c.Infrastructure = &Infrastructure{
		PgConnector: pgConnector,
		Logger:      logger,
		Metrics:     promMetrics,
	}

	return nil
}

func (c *Container) initQueries() error {
	c.Queries = &Queries{
		Slow: queries.NewSlowQueryHandler(c.Infrastructure.PgConnector.ReadPool),
	}

	return nil
}

func (c *Container) initMiddlewares() error {
	c.Middlewares = &Middlewares{
		Metrics: middlewares.NewMetricsMiddleware(c.Infrastructure.Metrics, c.Infrastructure.PgConnector.ReadPool, c.Infrastructure.PgConnector.WritePool),
	}

	return nil
}

func (c *Container) initHandlers() error {
	c.Handlers = &Handlers{
		Target:  handlers.NewTargetHandler(c.Infrastructure.Logger, c.Queries.Slow),
		Metrics: promhttp.Handler(),
	}

	return nil
}
