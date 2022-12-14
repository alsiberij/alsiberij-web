package main

import (
	"auth/internal/app"
	"auth/pkg/logging"
	"auth/pkg/pgs"
	"auth/pkg/rds"
	"crypto/tls"
	"encoding/json"
	"log"
	"os"
	"time"
)

type (
	Config struct {
		PgsAuth pgs.PostgresConfig `json:"pgs-1/auth"`
		Rds0    rds.RedisConfig    `json:"rds-1/0"`
		Rds1    rds.RedisConfig    `json:"rds-1/1"`
	}
)

func ReadConfig(filename string) (Config, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0444)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	_ = f.Close()
	return config, err
}

func main() {
	config, err := ReadConfig("./config.json")
	if err != nil {
		log.Fatalf("UNABLE READ CONFIG: %v", err)
	}

	pgsAuth, err := pgs.NewPostgres(config.PgsAuth)
	if err != nil {
		log.Fatalf("UNABLE CONNECT TO POSTGRES: %v", err)
	}
	defer pgsAuth.Close()

	rds0, err := rds.NewRedis(config.Rds0)
	if err != nil {
		log.Fatalf("UNABLE CONNECT TO REDIS0: %v", err)
	}
	defer rds0.Close()

	rds1, err := rds.NewRedis(config.Rds1)
	if err != nil {
		log.Fatalf("UNABLE CONNECT TO REDIS1: %v", err)
	}
	defer rds1.Close()

	logsPath := os.Getenv("LOGS_PATH")
	if logsPath == "" {
		logsPath = "./logs"
	}

	l := logging.NewLogger(1_000_000, logsPath+"/logs-%s.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777, "2006-01-02T15:04:05", time.Second*15)

	port := os.Getenv("PORT")
	if port == "" {
		port = "11400"
	}
	log.Printf("LISTENING %s PORT\n", port)

	sslPath := os.Getenv("SSL_PATH")
	if sslPath == "" {
		sslPath = "./../_ssl"
	}

	cert, err := tls.LoadX509KeyPair(sslPath+"/fullchain.pem", sslPath+"/privkey.pem")
	if err != nil {
		log.Fatalf("SSL ERROR: %v", err)
	}

	lis, err := tls.Listen("tcp4", ":"+port, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		log.Fatalf("LISTENER ERROR: %v", err)
	}

	srv, err := app.NewApp("API-GO-AUTH", l, pgsAuth, rds0, rds1, lis)
	if err != nil {
		log.Fatalf("APP ERROR: %v", err)
	}

	srv.Serve()

	log.Printf("BYE...\n")
}
