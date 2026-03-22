package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"laundry-backend/internal/config"
	"laundry-backend/internal/db"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("resetdb start")

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

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}
	if err := mustExec(ctx, pool, `SELECT 1`); err != nil {
		log.Fatalf("db query failed: %v", err)
	}

	log.Printf("drop schema laundry_backend (cascade)")
	if err := mustExec(ctx, pool, `DROP SCHEMA IF EXISTS laundry_backend CASCADE`); err != nil {
		log.Fatalf("drop schema failed: %v", err)
	}
	if err := mustExec(ctx, pool, `CREATE SCHEMA IF NOT EXISTS laundry_backend`); err != nil {
		log.Fatalf("create schema failed: %v", err)
	}

	log.Printf("run migrate")
	if err := runGo(ctx, "./cmd/migrate"); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	log.Printf("run seed")
	if err := runGo(ctx, "./cmd/seed"); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Printf("integrity checks")
	if err := integrityChecks(ctx, pool); err != nil {
		log.Fatalf("integrity checks failed: %v", err)
	}

	log.Printf("resetdb done")
}

func runGo(ctx context.Context, pkg string) error {
	cmd := exec.CommandContext(ctx, "go", "run", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func mustExec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) error {
	_, err := pool.Exec(ctx, sql, args...)
	return err
}

func count(ctx context.Context, pool *pgxpool.Pool, name, sql string) (int, error) {
	var n int
	if err := pool.QueryRow(ctx, sql).Scan(&n); err != nil {
		return 0, err
	}
	log.Printf("count %s=%d", name, n)
	return n, nil
}

func integrityChecks(ctx context.Context, pool *pgxpool.Pool) error {
	type check struct {
		name string
		sql  string
		min  int
	}

	counts := []check{
		{name: "users", sql: `SELECT COUNT(*) FROM laundry_backend.users`, min: 1},
		{name: "customers", sql: `SELECT COUNT(*) FROM laundry_backend.customers`, min: 1},
		{name: "employees", sql: `SELECT COUNT(*) FROM laundry_backend.employees`, min: 1},
		{name: "service_types", sql: `SELECT COUNT(*) FROM laundry_backend.service_types`, min: 1},
		{name: "orders", sql: `SELECT COUNT(*) FROM laundry_backend.orders`, min: 1},
		{name: "order_items", sql: `SELECT COUNT(*) FROM laundry_backend.order_items`, min: 1},
	}

	for _, c := range counts {
		n, err := count(ctx, pool, c.name, c.sql)
		if err != nil {
			return err
		}
		if n < c.min {
			return fmt.Errorf("table %s is empty (count=%d)", c.name, n)
		}
	}

	orphanChecks := []struct {
		name string
		sql  string
	}{
		{
			name: "order.customer orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.orders o
				LEFT JOIN laundry_backend.customers c ON c.id=o.customer_id
				WHERE c.id IS NULL
			`,
		},
		{
			name: "order_items.order orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.order_items oi
				LEFT JOIN laundry_backend.orders o ON o.id=oi.order_id
				WHERE o.id IS NULL
			`,
		},
		{
			name: "order_items.service_type orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.order_items oi
				LEFT JOIN laundry_backend.service_types st ON st.id=oi.service_type_id
				WHERE st.id IS NULL
			`,
		},
		{
			name: "payments.order orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.payments p
				LEFT JOIN laundry_backend.orders o ON o.id=p.order_id
				WHERE o.id IS NULL
			`,
		},
		{
			name: "work_assignments.order orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.work_assignments wa
				LEFT JOIN laundry_backend.orders o ON o.id=wa.order_id
				WHERE o.id IS NULL
			`,
		},
		{
			name: "work_assignments.order_item orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.work_assignments wa
				LEFT JOIN laundry_backend.order_items oi ON oi.id=wa.order_item_id
				WHERE oi.id IS NULL
			`,
		},
		{
			name: "work_assignments.employee orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.work_assignments wa
				LEFT JOIN laundry_backend.employees e ON e.id=wa.employee_id
				WHERE e.id IS NULL
			`,
		},
		{
			name: "order_attachments.order orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.order_attachments a
				LEFT JOIN laundry_backend.orders o ON o.id=a.order_id
				WHERE o.id IS NULL
			`,
		},
		{
			name: "delivery_stops.plan orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.delivery_stops s
				LEFT JOIN laundry_backend.delivery_plans p ON p.id=s.plan_id
				WHERE p.id IS NULL
			`,
		},
		{
			name: "delivery_stops.customer orphan",
			sql: `
				SELECT COUNT(*)
				FROM laundry_backend.delivery_stops s
				LEFT JOIN laundry_backend.customers c ON c.id=s.customer_id
				WHERE c.id IS NULL
			`,
		},
	}

	for _, oc := range orphanChecks {
		n, err := count(ctx, pool, oc.name, oc.sql)
		if err != nil {
			return err
		}
		if n != 0 {
			return fmt.Errorf("integrity error: %s count=%d", oc.name, n)
		}
	}

	return nil
}
