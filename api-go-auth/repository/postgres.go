package repository

import (
	"context"
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

	PostgresRepository struct {
		config PostgresConfig

		connPool *pgxpool.Pool
	}
)

var (
	AuthPostgresRepository PostgresRepository
)

func (r *PostgresRepository) Init(config PostgresConfig) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		config.User, config.Password, config.Host, config.Port, config.DbName, config.SslMode, config.MaxCons)
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return err
	}

	r.connPool = pool

	return nil
}

func (r *PostgresRepository) AcquireConnection() (*pgxpool.Conn, error) {
	return r.connPool.Acquire(context.Background())
}

func (r *PostgresRepository) UserRepository(q pgxtype.Querier) UserRepository {
	return &UserPostgresRepository{conn: q}
}

func (r *PostgresRepository) RefreshTokenRepository(q pgxtype.Querier) RefreshTokenRepository {
	return &RefreshTokenPostgresRepository{conn: q}
}
