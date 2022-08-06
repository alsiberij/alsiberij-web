package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"time"
)

const (
	ReconnectTimes = 3
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
	var i int
	var err error
	var pool *pgxpool.Pool

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		config.User, config.Password, config.Host, config.Port, config.DbName, config.SslMode, config.MaxCons)

	for ; i < ReconnectTimes; i++ {
		log.Printf("CONNECTING #%d [%s]...\n", i+1, dsn)
		pool, err = pgxpool.Connect(context.Background(), dsn)
		if err != nil {
			log.Printf("CONNECTION #%d FAILED WITH ERROR: %s\n", i+1, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}

		err = pool.Ping(context.Background())
		if err != nil {
			log.Printf("CONNECTION #%d FAILED WITH ERROR: %s\n", i+1, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	if i == ReconnectTimes {
		log.Println("ALL RETRIES FAILED")
		return err
	}

	log.Printf("CONNECTION #%d SUCCED\n", i)
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
