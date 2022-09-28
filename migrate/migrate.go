package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"migrate/pkg/pgs"
	"os"
)

type (
	Config struct {
		PgsAuth pgs.PostgresConfig `json:"pgs-1/auth"`
	}
)

var (
	migrateUp    bool
	migrateDown  bool
	migrateSteps int64
)

func init() {
	flag.BoolVar(&migrateUp, "up", false, "-up: apply all migrations")
	flag.BoolVar(&migrateDown, "down", false, "-up: rollback all migrations")
	flag.Int64Var(&migrateSteps, "steps", 0, "-steps: apply/rollback N migrations")
	flag.Parse()

	if !migrateUp && !migrateDown && migrateSteps == 0 {
		log.Fatal("ARGUMENTS ERROR: PROVIDE -up/-down/-steps")
	}

	if migrateUp && migrateDown {
		log.Fatal("ARGUMENTS ERROR: ONLY ONE FROM -up/-down IS ALLOWED")
	}
}

func main() {
	f, err := os.OpenFile("./config.json", os.O_RDONLY, 0444)
	if err != nil {
		log.Fatalf("CONFIG ERROR: %v", err)
	}

	var cfg Config
	err = json.NewDecoder(f).Decode(&cfg)
	_ = f.Close()
	if err != nil {
		log.Fatalf("CONFIG ERROR: %v", err)
	}

	m, err := migrate.New("file:///migrations/api-go-auth/", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.PgsAuth.User, cfg.PgsAuth.Password,
		cfg.PgsAuth.Host, cfg.PgsAuth.Port,
		cfg.PgsAuth.DbName, cfg.PgsAuth.SslMode))
	if err != nil {
		log.Fatalf("DB ERROR: %v", err)
	}

	if migrateUp {
		err = m.Up()
	} else if migrateDown {
		err = m.Down()
	} else {
		err = m.Steps(int(migrateSteps))
	}

	if err == migrate.ErrNoChange {
		err = nil
	}

	if err != nil {
		log.Fatalf("MIGRATION ERROR: %v", err)
	}

	log.Println("MIGRATE OK")
}
