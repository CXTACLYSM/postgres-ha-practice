package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	WriteOperation = 1
	ReadOperation  = 2
)

type Connector struct {
	ReadPool  *pgxpool.Pool
	WritePool *pgxpool.Pool
}

func NewConnector(readDSN, writeDSN string) (*Connector, error) {
	readPool, err := getPool(readDSN)
	if err != nil {
		return nil, err
	}

	writePool, err := getPool(writeDSN)
	if err != nil {
		return nil, err
	}

	return &Connector{
		ReadPool:  readPool,
		WritePool: writePool,
	}, nil
}

func getPool(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	poolCfg.MaxConns = 50
	poolCfg.MinConns = 5
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("cannot connect to PostgreSQL: %w", err)
	}
	log.Printf("successfully connected to PostgreSQL at %s\n", dsn)

	return pool, nil
}

func (c *Connector) Close() {
	c.ReadPool.Close()
	c.WritePool.Close()
}

func (c *Connector) PingByOperation(ctx context.Context, operation uint8) error {
	switch true {
	case operation == ReadOperation && c.ReadPool != nil:
		err := c.ReadPool.Ping(ctx)
		if err != nil {
			return err
		}
		return nil
	case operation == WriteOperation && c.WritePool != nil:
		err := c.WritePool.Ping(ctx)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("operation not supported or corresponding pool is nil")
	}
}
