package db

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBClient struct {
	Pool *pgxpool.Pool
}

//go:embed queries/insert_result.pgsql
var insertResultQuery string

func NewDBClient(ctx context.Context, connString string) (*DBClient, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBClient{Pool: pool}, nil
}

func (c *DBClient) Close() {
	c.Pool.Close()
}

type Result struct {
	Name    string
	Nonce   int
	Sha256  string
	Quality float64
}

func (c *DBClient) InsertResult(
	ctx context.Context,
	name string,
	nonce string,
	sha256 string,
	quality float64,
) error {
	_, err := c.Pool.Exec(ctx, insertResultQuery, name, nonce, sha256, quality)
	if err != nil {
		return fmt.Errorf("insert link: %w", err)
	}

	return nil
}
