package database

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

	Postgres struct {
		conn *pgxpool.Pool
	}
)

func NewPostgres(config PostgresConfig) (Postgres, error) {
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

	return Postgres{conn: pool}, nil
}

func (r *Postgres) AcquireConnection() (*pgxpool.Conn, error) {
	return r.conn.Acquire(context.Background())
}

func (r *Postgres) Users(q pgxtype.Querier) Users {
	return Users{conn: q}
}

func (r *Postgres) RefreshTokens(q pgxtype.Querier) RefreshTokens {
	return RefreshTokens{conn: q}
}
