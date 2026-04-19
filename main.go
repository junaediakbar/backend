package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"laundry-backend/internal/config"
	"laundry-backend/internal/db"
	httpserver "laundry-backend/internal/http"
	"laundry-backend/internal/http/handler"
	"laundry-backend/internal/http/middleware"
	"laundry-backend/internal/migrate"
	"laundry-backend/internal/repository/pg"
	"laundry-backend/internal/service"
)

func main() {
	cfg := config.Load()
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Avoid uploads/ piling up on ephemeral environments.
	_ = os.RemoveAll("uploads")
	_ = os.RemoveAll("./uploads")

	log.Printf("starting server addr=%s", cfg.HTTPAddr)

	pool, err := db.NewPool(ctx, db.PoolConfig{
		DatabaseURL:     cfg.DatabaseURL,
		MaxConns:        cfg.DBMaxConns,
		MinConns:        cfg.DBMinConns,
		MaxConnIdleTime: cfg.DBMaxConnIdleTime,
		MaxConnLifetime: cfg.DBMaxConnLifetime,
		HealthCheck:     cfg.DBHealthCheck,
	})
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		log.Printf("running migrations dir=%s", cfg.MigrationsDir)
		if err := migrate.Up(ctx, pool, cfg.MigrationsDir); err != nil {
			panic(err)
		}
		log.Printf("migrations done")
	}

	dbpg := pg.New(pool)

	customerRepo := pg.NewCustomerRepo(dbpg)
	serviceTypeRepo := pg.NewServiceTypeRepo(dbpg)
	employeeRepo := pg.NewEmployeeRepo(dbpg)
	orderRepo := pg.NewOrderRepo(dbpg, cfg.Timezone)
	deliveryRepo := pg.NewDeliveryRepo(dbpg)
	dashboardRepo := pg.NewDashboardRepo(dbpg)
	reportRepo := pg.NewReportRepo(dbpg)

	customerSvc := service.NewCustomerService(customerRepo)
	serviceTypeSvc := service.NewServiceTypeService(serviceTypeRepo)
	employeeSvc := service.NewEmployeeService(employeeRepo)
	orderSvc := service.NewOrderService(orderRepo)
	deliverySvc := service.NewDeliveryService(deliveryRepo)
	dashboardSvc := service.NewDashboardService(dashboardRepo)
	reportSvc := service.NewReportService(reportRepo)
	authSvc := service.NewAuthService(employeeRepo)

	router := httpserver.NewRouter(httpserver.ServerDeps{
		Auth: middleware.AuthConfig{
			Mode:           cfg.AuthMode,
			APIKey:         cfg.APIKey,
			SupabaseJWKS:   cfg.SupabaseJWKSURL,
			SupabaseIssuer: cfg.SupabaseIssuer,
			JWTSecret:      cfg.JWTSecret,
		},
		Authn:          handler.NewAuthHandler(authSvc, cfg.JWTSecret),
		PublicReceipts: handler.NewPublicReceiptHandler(orderRepo, cfg.Timezone),
		Dashboard:      handler.NewDashboardHandler(dashboardSvc, cfg.Timezone),
		Customers:      handler.NewCustomerHandler(customerSvc),
		Orders:         handler.NewOrderHandler(orderSvc, cfg.Timezone),
		ServiceTypes:   handler.NewServiceTypeHandler(serviceTypeSvc),
		Employees:      handler.NewEmployeeHandler(employeeSvc, cfg.Timezone),
		Delivery:       handler.NewDeliveryHandler(deliverySvc),
		Reports:        handler.NewReportHandler(reportSvc, cfg.Timezone),
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("listening addr=%s", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
