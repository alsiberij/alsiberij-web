package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
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
		connPool *pgxpool.Pool
		isActive bool
	}
)

var (
	errPostgresNotInitialized = errors.New("postgres not initialized")
)

func New(config PostgresConfig) (Postgres, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		config.User, config.Password, config.Host, config.Port, config.DbName, config.SslMode, config.MaxCons)

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return Postgres{}, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return Postgres{}, err
	}

	return Postgres{connPool: pool, isActive: true}, nil
}

func (r *Postgres) AcquireConnection() (*pgxpool.Conn, error) {
	if !r.isActive {
		return nil, errPostgresNotInitialized
	}
	return r.connPool.Acquire(context.Background())
}

func (r *Postgres) Users(q pgxtype.Querier) Users {
	return &UsersPostgres{conn: q}
}

func (r *Postgres) RefreshTokens(q pgxtype.Querier) RefreshTokens {
	return &RefreshTokensPostgres{conn: q}
}
