package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"laundry-backend/internal/config"
	"laundry-backend/internal/db"
	"laundry-backend/internal/migrate"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("migrate start dir=%s", cfg.MigrationsDir)
	pool, err := db.NewPool(ctx, db.PoolConfig{
		DatabaseURL:     cfg.DatabaseURL,
		MaxConns:        cfg.DBMaxConns,
		MinConns:        cfg.DBMinConns,
		MaxConnIdleTime: cfg.DBMaxConnIdleTime,
		MaxConnLifetime: cfg.DBMaxConnLifetime,
		HealthCheck:     cfg.DBHealthCheck,
	})
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, cfg.MigrationsDir); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}
	log.Printf("migrate done")
}
