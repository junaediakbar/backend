package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lucsky/cuid"
	"golang.org/x/crypto/bcrypt"

	"laundry-backend/internal/config"
	"laundry-backend/internal/db"
)

type serviceTypeSeed struct {
	name         string
	unit         string
	defaultPrice string
}

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	log.Printf("seed start")
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

	if err := seedServiceTypes(ctx, pool); err != nil {
		log.Fatalf("seed service types failed: %v", err)
	}
	if err := seedEmployees(ctx, pool, rng, 10); err != nil {
		log.Fatalf("seed employees failed: %v", err)
	}
	if err := seedCustomers(ctx, pool, rng, 10); err != nil {
		log.Fatalf("seed customers failed: %v", err)
	}
	if err := seedOrders(ctx, pool, rng, 50); err != nil {
		log.Fatalf("seed orders failed: %v", err)
	}
	if err := seedAdminUser(ctx, pool); err != nil {
		log.Fatalf("seed admin user failed: %v", err)
	}
	if err := seedDefaultUsers(ctx, pool); err != nil {
		log.Fatalf("seed default users failed: %v", err)
	}

	log.Printf("seed done")
}

func seedAdminUser(ctx context.Context, pool *pgxpool.Pool) error {
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || strings.TrimSpace(password) == "" {
		log.Printf("seed admin user skipped (ADMIN_EMAIL/ADMIN_PASSWORD empty)")
		return nil
	}

	var exists int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.users WHERE lower(email)=lower($1)`, email).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		log.Printf("seed admin user skipped (already exists)")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	id := cuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO laundry_backend.users (id, auth_user_id, name, email, role, password_hash, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,'owner',$5,true,now(),now())
	`, id, id, "Owner", email, string(hash))
	if err != nil {
		return err
	}

	log.Printf("seed admin user created email=%s role=owner", email)
	return nil
}

func seedDefaultUsers(ctx context.Context, pool *pgxpool.Pool) error {
	type row struct {
		envEmail    string
		envPassword string
		name        string
		role        string
	}

	users := []row{
		{envEmail: "SEED_ADMIN_EMAIL", envPassword: "SEED_ADMIN_PASSWORD", name: "Admin", role: "admin"},
		{envEmail: "SEED_CASHIER_EMAIL", envPassword: "SEED_CASHIER_PASSWORD", name: "Cashier", role: "cashier"},
	}

	created := 0
	for _, u := range users {
		email := strings.TrimSpace(os.Getenv(u.envEmail))
		password := os.Getenv(u.envPassword)
		if email == "" || strings.TrimSpace(password) == "" {
			continue
		}
		if err := seedUser(ctx, pool, u.name, email, u.role, password); err != nil {
			return err
		}
		created++
	}

	log.Printf("seed default users created=%d", created)
	return nil
}

func seedUser(ctx context.Context, pool *pgxpool.Pool, name, email, role, password string) error {
	var exists int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.users WHERE lower(email)=lower($1)`, email).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	id := cuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO laundry_backend.users (id, auth_user_id, name, email, role, password_hash, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,true,now(),now())
	`, id, id, name, email, role, string(hash))
	return err
}

func seedServiceTypes(ctx context.Context, pool *pgxpool.Pool) error {
	defaults := []serviceTypeSeed{
		{name: "Karpet Malaysia Tipis", unit: "m2", defaultPrice: "20000.00"},
		{name: "Karpet Malaysia Tebal", unit: "m2", defaultPrice: "25000.00"},
		{name: "Karpet Permadani Tipis", unit: "m2", defaultPrice: "15000.00"},
		{name: "Karpet Permadani Tebal", unit: "m2", defaultPrice: "18000.00"},
		{name: "Kasur Karakter Tipis", unit: "m2", defaultPrice: "25000.00"},
		{name: "Kasur Karakter Tebal", unit: "m2", defaultPrice: "30000.00"},
		{name: "Kasur Bulu Tipis", unit: "m2", defaultPrice: "25000.00"},
		{name: "Kasur Bulu Tebal", unit: "m2", defaultPrice: "30000.00"},
		{name: "Kasur Bulu Super Tebal", unit: "m2", defaultPrice: "35000.00"},
		{name: "Karpet Rol Polos Tipis", unit: "m2", defaultPrice: "12000.00"},
		{name: "Karpet Rol Tebal/ blk anyam", unit: "m2", defaultPrice: "15000.00"},
		{name: "Karpet Masjid Tipis", unit: "m2", defaultPrice: "18000.00"},
		{name: "Karpet Masjid Tebal", unit: "m2", defaultPrice: "22000.00"},
		{name: "Karpet Masjid Super Tebal", unit: "m2", defaultPrice: "30000.00"},
		{name: "Karpet Turki Tipis", unit: "m2", defaultPrice: "20000.00"},
		{name: "Karpet Turki Tebal", unit: "m2", defaultPrice: "25000.00"},
		{name: "Karpet Bulu Tipis", unit: "m2", defaultPrice: "17000.00"},
		{name: "Karpet Bulu Tebal", unit: "m2", defaultPrice: "20000.00"},
		{name: "Ambal", unit: "m2", defaultPrice: "10000.00"},
		{name: "Rumbai diputihkan/ dibersihkan", unit: "m1", defaultPrice: "2000.00"},
	}

	var inserted int
	for _, s := range defaults {
		_, err := pool.Exec(ctx, `
			INSERT INTO laundry_backend.service_types (id, name, unit, default_price, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, true, now(), now())
			ON CONFLICT (name) DO UPDATE SET
				unit = EXCLUDED.unit,
				default_price = EXCLUDED.default_price,
				is_active = true,
				updated_at = now()
		`, cuid.New(), s.name, s.unit, s.defaultPrice)
		if err != nil {
			return err
		}
		inserted++
	}

	log.Printf("seed service types upserted=%d", inserted)
	return nil
}

func seedEmployees(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, target int) error {
	var existing int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.employees WHERE name ILIKE 'Demo %'`).Scan(&existing); err != nil {
		return err
	}
	if existing >= target {
		log.Printf("seed employees skipped existing=%d target=%d", existing, target)
		return nil
	}

	firstNames := []string{"Ari", "Budi", "Dewi", "Fitri", "Hendra", "Iwan", "Joko", "Maya", "Nina", "Rizky", "Sari", "Tono"}
	roles := []string{"Driver", "Operator", "Admin", "Packing"}

	toInsert := target - existing
	for i := 0; i < toInsert; i++ {
		name := fmt.Sprintf("Demo %s %s %02d", roles[rng.Intn(len(roles))], firstNames[rng.Intn(len(firstNames))], existing+i+1)
		_, err := pool.Exec(ctx, `
			INSERT INTO laundry_backend.employees (id, name, is_active, created_at, updated_at)
			VALUES ($1, $2, true, now(), now())
		`, cuid.New(), name)
		if err != nil {
			return err
		}
	}

	log.Printf("seed employees inserted=%d", toInsert)
	return nil
}

type customerSeed struct {
	name    string
	phone   string
	address string
	email   string
	lat     string
	lng     string
}

func seedCustomers(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, target int) error {
	var existing int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.customers WHERE name ILIKE 'Demo %'`).Scan(&existing); err != nil {
		return err
	}
	if existing >= target {
		log.Printf("seed customers skipped existing=%d target=%d", existing, target)
		return nil
	}

	baseLat := -6.200000
	baseLng := 106.816666

	toInsert := target - existing
	for i := 0; i < toInsert; i++ {
		n := existing + i + 1
		lat := baseLat + (rng.Float64()-0.5)*0.2
		lng := baseLng + (rng.Float64()-0.5)*0.2
		c := customerSeed{
			name:    fmt.Sprintf("Demo Pelanggan %02d", n),
			phone:   fmt.Sprintf("08%02d%08d", rng.Intn(90)+10, rng.Intn(100000000)),
			address: fmt.Sprintf("Demo Address %02d, Jakarta", n),
			email:   fmt.Sprintf("demo%02d@example.com", n),
			lat:     fmt.Sprintf("%.6f", lat),
			lng:     fmt.Sprintf("%.6f", lng),
		}
		_, err := pool.Exec(ctx, `
			INSERT INTO laundry_backend.customers (id, name, phone, address, latitude, longitude, email, notes, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now(),now())
		`, cuid.New(), c.name, c.phone, c.address, c.lat, c.lng, c.email, "DEMO")
		if err != nil {
			return err
		}
	}

	log.Printf("seed customers inserted=%d", toInsert)
	return nil
}

type serviceTypeRow struct {
	id           string
	name         string
	unit         string
	defaultPrice string
}

func seedOrders(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, target int) error {
	var existing int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.orders WHERE invoice_number LIKE 'DEMO-%'`).Scan(&existing); err != nil {
		return err
	}
	if existing >= target {
		if err := backfillDemoOrderDetails(ctx, pool, rng, target); err != nil {
			return err
		}
		log.Printf("seed orders skipped existing=%d target=%d (backfill applied)", existing, target)
		return nil
	}

	customerIDs, err := loadIDs(ctx, pool, `SELECT id FROM laundry_backend.customers ORDER BY created_at ASC`)
	if err != nil {
		return err
	}
	if len(customerIDs) == 0 {
		return fmt.Errorf("no customers found")
	}

	serviceTypes, err := loadServiceTypes(ctx, pool)
	if err != nil {
		return err
	}
	if len(serviceTypes) == 0 {
		return fmt.Errorf("no service types found")
	}

	employeeIDs, err := loadIDs(ctx, pool, `SELECT id FROM laundry_backend.employees WHERE is_active=true ORDER BY created_at ASC`)
	if err != nil {
		return err
	}

	toInsert := target - existing
	for i := 0; i < toInsert; i++ {
		orderID := cuid.New()
		invoice := fmt.Sprintf("DEMO-%s-%03d", time.Now().Format("20060102"), existing+i+1)
		customerID := customerIDs[rng.Intn(len(customerIDs))]

		receivedAt := time.Now().Add(-time.Duration(rng.Intn(30*24)) * time.Hour)
		workflow := []string{"received", "washing", "drying", "ironing", "finished", "picked_up"}[rng.Intn(6)]

		itemCount := rng.Intn(3) + 1
		items := make([]orderItemSeed, 0, itemCount)
		noteParts := make([]string, 0, itemCount)
		var total float64
		for j := 0; j < itemCount; j++ {
			st := serviceTypes[rng.Intn(len(serviceTypes))]
			qty, sizeNote := randomQuantity(rng, st.unit)
			unitPrice := parseMoney(st.defaultPrice)
			discount := 0.0
			if rng.Intn(10) == 0 {
				discount = float64((rng.Intn(5) + 1) * 1000)
			}
			lineTotal := qty*unitPrice - discount
			if lineTotal < 0 {
				lineTotal = 0
			}
			total += lineTotal
			items = append(items, orderItemSeed{
				id:          cuid.New(),
				serviceType: st,
				quantity:    fmt.Sprintf("%.2f", qty),
				unitPrice:   fmt.Sprintf("%.2f", unitPrice),
				discount:    fmt.Sprintf("%.2f", discount),
				total:       fmt.Sprintf("%.2f", lineTotal),
				totalValue:  lineTotal,
			})
			if sizeNote != "" {
				noteParts = append(noteParts, fmt.Sprintf("%s: %s", st.name, sizeNote))
			}
		}

		paymentStatus := "unpaid"
		var paymentAmount float64
		if rng.Intn(100) < 65 {
			if rng.Intn(100) < 20 {
				paymentStatus = "partial"
				paymentAmount = total * (0.3 + rng.Float64()*0.4)
			} else {
				paymentStatus = "paid"
				paymentAmount = total
			}
		}

		var note *string
		if len(noteParts) > 0 {
			s := "DEMO " + joinLines(noteParts)
			note = &s
		}

		if err := insertOrderTx(ctx, pool, rng, employeeIDs, orderID, invoice, customerID, paymentStatus, workflow, receivedAt, total, note, items, paymentAmount); err != nil {
			return err
		}
	}

	if err := backfillDemoOrderDetails(ctx, pool, rng, target); err != nil {
		return err
	}

	log.Printf("seed orders inserted=%d", toInsert)
	return nil
}

type orderItemSeed struct {
	id          string
	serviceType serviceTypeRow
	quantity    string
	unitPrice   string
	discount    string
	total       string
	totalValue  float64
}

func insertOrderTx(
	ctx context.Context,
	pool *pgxpool.Pool,
	rng *rand.Rand,
	employeeIDs []string,
	orderID, invoice, customerID, paymentStatus, workflowStatus string,
	receivedAt time.Time,
	total float64,
	note *string,
	items []orderItemSeed,
	paymentAmount float64,
) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO laundry_backend.orders (
			id,
			invoice_number,
			customer_id,
			total,
			payment_status,
			workflow_status,
			received_date,
			completed_date,
			pickup_date,
			note,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NULL,NULL,$8,$7,$7)
	`, orderID, invoice, customerID, fmt.Sprintf("%.2f", total), paymentStatus, workflowStatus, receivedAt, note)
	if err != nil {
		return err
	}

	for _, it := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO laundry_backend.order_items (
				id,
				order_id,
				service_type_id,
				quantity,
				unit_price,
				discount,
				total,
				created_at,
				updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)
		`, it.id, orderID, it.serviceType.id, it.quantity, it.unitPrice, it.discount, it.total, receivedAt)
		if err != nil {
			return err
		}
	}

	if len(employeeIDs) > 0 {
		for _, it := range items {
			if err := seedWorkAssignmentsTx(ctx, tx, rng, employeeIDs, orderID, it.id, it.totalValue, receivedAt); err != nil {
				return err
			}
		}
	}

	if rng.Intn(100) < 35 {
		if err := seedOrderAttachmentTx(ctx, tx, orderID, receivedAt); err != nil {
			return err
		}
	}

	if paymentAmount > 0 {
		_, err := tx.Exec(ctx, `
			INSERT INTO laundry_backend.payments (id, order_id, amount, method, paid_at, note, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$5)
		`, cuid.New(), orderID, fmt.Sprintf("%.2f", paymentAmount), "cash", receivedAt, "DEMO")
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func backfillDemoOrderDetails(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, limit int) error {
	orderIDs, err := loadDemoOrderIDs(ctx, pool, limit)
	if err != nil {
		return err
	}
	if len(orderIDs) == 0 {
		return nil
	}

	employeeIDs, err := loadIDs(ctx, pool, `SELECT id FROM laundry_backend.employees WHERE is_active=true ORDER BY created_at ASC`)
	if err != nil {
		return err
	}

	for _, orderID := range orderIDs {
		if err := backfillDemoOrderDetail(ctx, pool, rng, employeeIDs, orderID); err != nil {
			return err
		}
	}
	return nil
}

func backfillDemoOrderDetail(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, employeeIDs []string, orderID string) error {
	var waCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.work_assignments WHERE order_id=$1`, orderID).Scan(&waCount); err != nil {
		return err
	}
	var attCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM laundry_backend.order_attachments WHERE order_id=$1`, orderID).Scan(&attCount); err != nil {
		return err
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var createdAt time.Time
	if err := tx.QueryRow(ctx, `SELECT created_at FROM laundry_backend.orders WHERE id=$1`, orderID).Scan(&createdAt); err != nil {
		return err
	}

	if waCount == 0 && len(employeeIDs) > 0 {
		rows, err := tx.Query(ctx, `SELECT id, total::float8 FROM laundry_backend.order_items WHERE order_id=$1 ORDER BY created_at ASC`, orderID)
		if err != nil {
			return err
		}
		type itemRow struct {
			id    string
			total float64
		}
		items := []itemRow{}
		for rows.Next() {
			var it itemRow
			if err := rows.Scan(&it.id, &it.total); err != nil {
				rows.Close()
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		for _, it := range items {
			if err := seedWorkAssignmentsTx(ctx, tx, rng, employeeIDs, orderID, it.id, it.total, createdAt); err != nil {
				return err
			}
		}
	}

	if attCount == 0 && rng.Intn(100) < 35 {
		if err := seedOrderAttachmentTx(ctx, tx, orderID, createdAt); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func loadDemoOrderIDs(ctx context.Context, pool *pgxpool.Pool, limit int) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT id
		FROM laundry_backend.orders
		WHERE invoice_number LIKE 'DEMO-%'
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func seedWorkAssignmentsTx(ctx context.Context, tx pgx.Tx, rng *rand.Rand, employeeIDs []string, orderID, orderItemID string, itemTotal float64, createdAt time.Time) error {
	type task struct {
		taskType string
		percent  float64
	}
	tasks := []task{
		{taskType: "pickup", percent: 5},
		{taskType: "finishing_packing", percent: 10},
	}
	for _, t := range tasks {
		employeeID := employeeIDs[rng.Intn(len(employeeIDs))]
		amount := round2(itemTotal * t.percent / 100)
		_, err := tx.Exec(ctx, `
			INSERT INTO laundry_backend.work_assignments (id, order_id, order_item_id, employee_id, task_type, percent, amount, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (order_item_id, task_type) DO NOTHING
		`, cuid.New(), orderID, orderItemID, employeeID, t.taskType, fmt.Sprintf("%.2f", t.percent), fmt.Sprintf("%.2f", amount), createdAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedOrderAttachmentTx(ctx context.Context, tx pgx.Tx, orderID string, createdAt time.Time) error {
	fileURL := fmt.Sprintf("https://picsum.photos/seed/%s/800/600", orderID)
	_, err := tx.Exec(ctx, `
		INSERT INTO laundry_backend.order_attachments (id, order_id, file_path, mime_type, size_bytes, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, cuid.New(), orderID, fileURL, "image/jpeg", 0, createdAt)
	return err
}

func loadIDs(ctx context.Context, pool *pgxpool.Pool, q string) ([]string, error) {
	rows, err := pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func loadServiceTypes(ctx context.Context, pool *pgxpool.Pool) ([]serviceTypeRow, error) {
	rows, err := pool.Query(ctx, `SELECT id, name, unit, default_price::text FROM laundry_backend.service_types WHERE is_active=true ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []serviceTypeRow
	for rows.Next() {
		var r serviceTypeRow
		if err := rows.Scan(&r.id, &r.name, &r.unit, &r.defaultPrice); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func randomQuantity(rng *rand.Rand, unit string) (float64, string) {
	switch unit {
	case "m2":
		panjang := float64(rng.Intn(4)+1) + float64(rng.Intn(10))/10
		lebar := float64(rng.Intn(3)+1) + float64(rng.Intn(10))/10
		return round2(panjang * lebar), fmt.Sprintf("%.1fx%.1fm", panjang, lebar)
	case "m1":
		q := float64(rng.Intn(10)+1) + float64(rng.Intn(10))/10
		return round2(q), fmt.Sprintf("%.1fm", q)
	default:
		return 1, ""
	}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func parseMoney(s string) float64 {
	var v float64
	_, _ = fmt.Sscanf(s, "%f", &v)
	return v
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	out := lines[0]
	for i := 1; i < len(lines); i++ {
		out += " | " + lines[i]
	}
	return out
}
