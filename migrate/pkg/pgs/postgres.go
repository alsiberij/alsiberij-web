package pgs

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

type (
	PostgresConfig struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		DbName   string `json:"dbName"`
		SslMode  string `json:"sslMode"`
		MaxCons  int    `json:"maxCons"`
	}

	Postgres struct {
		conn *pgxpool.Pool
	}
)

var (
	ErrNotInitialized = errors.New("nil db")
)

func NewPostgres(config PostgresConfig) (*Postgres, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		config.User, config.Password, config.Host, config.Port, config.DbName, config.SslMode, config.MaxCons)

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return &Postgres{conn: pool}, nil
}

func (r *Postgres) AcquireConnection(ctx context.Context) (*pgxpool.Conn, error) {
	if r.conn == nil {
		return nil, ErrNotInitialized
	}

	return r.conn.Acquire(ctx)
}

func (r *Postgres) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}
