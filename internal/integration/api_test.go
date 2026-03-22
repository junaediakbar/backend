package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"laundry-backend/internal/db"
	httpserver "laundry-backend/internal/http"
	"laundry-backend/internal/http/handler"
	"laundry-backend/internal/http/middleware"
	"laundry-backend/internal/migrate"
	"laundry-backend/internal/repository/pg"
	"laundry-backend/internal/service"
)

func TestAPI_CreateCustomerAndList(t *testing.T) {
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("testcontainers provider tidak tersedia: %v", r)
		}
	}()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Skipf("testcontainers provider tidak tersedia: %v", err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := db.NewPool(ctx, db.PoolConfig{
		DatabaseURL:     connStr,
		MaxConns:        4,
		MinConns:        1,
		MaxConnIdleTime: 1 * time.Minute,
		MaxConnLifetime: 5 * time.Minute,
		HealthCheck:     10 * time.Second,
	})
	require.NoError(t, err)
	defer pool.Close()

	wd, err := os.Getwd()
	require.NoError(t, err)
	migrationsDir := filepath.Clean(filepath.Join(wd, "..", "..", "migrations"))
	require.NoError(t, migrate.Up(ctx, pool, migrationsDir))

	dbpg := pg.New(pool)

	router := httpserver.NewRouter(httpserver.ServerDeps{
		Auth:         middleware.AuthConfig{Mode: "none"},
		Dashboard:    handler.NewDashboardHandler(service.NewDashboardService(pg.NewDashboardRepo(dbpg))),
		Customers:    handler.NewCustomerHandler(service.NewCustomerService(pg.NewCustomerRepo(dbpg))),
		Orders:       handler.NewOrderHandler(service.NewOrderService(pg.NewOrderRepo(dbpg))),
		ServiceTypes: handler.NewServiceTypeHandler(service.NewServiceTypeService(pg.NewServiceTypeRepo(dbpg))),
		Employees:    handler.NewEmployeeHandler(service.NewEmployeeService(pg.NewEmployeeRepo(dbpg))),
		Delivery:     handler.NewDeliveryHandler(service.NewDeliveryService(pg.NewDeliveryRepo(dbpg))),
		Reports:      handler.NewReportHandler(service.NewReportService(pg.NewReportRepo(dbpg))),
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	createBody := map[string]any{
		"name":      "Budi",
		"phone":     "08123456789",
		"address":   "Jl. Mawar",
		"latitude":  -6.2,
		"longitude": 106.8,
	}
	raw, _ := json.Marshal(createBody)
	res, err := http.Post(srv.URL+"/api/v1/customers", "application/json", bytes.NewReader(raw))
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)

	var created struct {
		OK   bool `json:"ok"`
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&created))
	require.True(t, created.OK)
	require.NotEmpty(t, created.Data.ID)
	require.Equal(t, "Budi", created.Data.Name)

	res2, err := http.Get(srv.URL + "/api/v1/customers?page=1&pageSize=20&q=Budi")
	require.NoError(t, err)
	defer res2.Body.Close()
	require.Equal(t, http.StatusOK, res2.StatusCode)

	var listed struct {
		OK   bool `json:"ok"`
		Data struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res2.Body).Decode(&listed))
	require.True(t, listed.OK)
	require.GreaterOrEqual(t, listed.Data.Total, 1)
	require.GreaterOrEqual(t, len(listed.Data.Items), 1)
}
