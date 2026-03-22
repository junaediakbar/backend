# Laundry Backend (Go)

Backend API untuk aplikasi Laundry (Go + PostgreSQL).

## Prasyarat

- Go 1.22+
- PostgreSQL (contoh: Neon / lokal)

## Konfigurasi Environment

Buat file `backend/.env` (atau set env vars di shell). Variabel yang dipakai:

```bash
# Server
HTTP_ADDR=:8080

# Database
DATABASE_URL=postgresql://USER:PASSWORD@HOST:5432/DB?sslmode=require
DB_MAX_CONNS=10
DB_MIN_CONNS=2
DB_MAX_CONN_IDLE_TIME=5m
DB_MAX_CONN_LIFETIME=30m
DB_HEALTH_CHECK=30s

# Auth
# AUTH_MODE: none | api_key | jwt | supabase_jwks
AUTH_MODE=jwt
API_KEY=CHANGE_ME_RANDOM   # dipakai jika AUTH_MODE=api_key
JWT_SECRET=CHANGE_ME_RANDOM_LONG # dipakai jika AUTH_MODE=jwt
SUPABASE_JWKS_URL=         # dipakai jika AUTH_MODE=supabase_jwks
SUPABASE_ISSUER=

# Migrations
RUN_MIGRATIONS=false
MIGRATIONS_DIR=./migrations

# Optional seed users (dipakai oleh cmd/seed)
ADMIN_EMAIL=owner@laundry.local
ADMIN_PASSWORD=123456
SEED_ADMIN_EMAIL=admin@laundry.local
SEED_ADMIN_PASSWORD=123456
SEED_CASHIER_EMAIL=cashier@laundry.local
SEED_CASHIER_PASSWORD=123456
```

## Menjalankan Server

```bash
cd backend
go run .
```

Healthcheck:

```bash
curl -s http://localhost:8080/health
```

Dokumentasi API:

- Swagger UI: `http://localhost:8080/docs`
- OpenAPI JSON: `http://localhost:8080/openapi.json`

## Migrations

Jalankan migrasi (membuat schema `laundry_backend` dan tabel-tabelnya):

```bash
cd backend
go run ./cmd/migrate
```

Jika ingin auto-run saat start server, set:

```bash
RUN_MIGRATIONS=true
```

## Seeding Data

Seeder akan mengisi data demo (service types, customers, employees, orders) dan user default dari env:

```bash
cd backend
go run ./cmd/seed
```

Catatan:

- Seed user `owner` akan dibuat dari `ADMIN_EMAIL`/`ADMIN_PASSWORD` jika belum ada.
- Seed user `admin` dan `cashier` dibuat dari `SEED_ADMIN_*` dan `SEED_CASHIER_*` (opsional).

## Autentikasi

Semua endpoint `/api/v1/*` (kecuali `/api/v1/auth/login`) diproteksi oleh middleware auth.

### Mode JWT (direkomendasikan)

Set:

```bash
AUTH_MODE=jwt
JWT_SECRET=...
```

Login:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@laundry.local","password":"123456"}'
```

Respon akan berisi `token`. Gunakan sebagai Bearer token:

```bash
curl -s http://localhost:8080/api/v1/dashboard/summary \
  -H "Authorization: Bearer <TOKEN>"
```

### Mode API Key

Set:

```bash
AUTH_MODE=api_key
API_KEY=...
```

Panggil API:

```bash
curl -s http://localhost:8080/api/v1/dashboard/summary \
  -H "X-API-Key: <API_KEY>"
```

## User & Role Management

Endpoint user management:

- `GET /api/v1/users` (owner/admin)
- `POST /api/v1/users` (owner/admin)
- `GET /api/v1/users/{id}` (owner/admin)
- `PUT /api/v1/users/{id}` (owner/admin)
- `DELETE /api/v1/users/{id}` (owner saja)
