package config

import (
	"bufio"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultDatabaseURL = "postgresql://neondb_owner:npg_acJFjY1Bum7E@ep-square-fog-a11iotje-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require"

type Config struct {
	HTTPAddr string

	DatabaseURL string
	DBMaxConns  int32
	DBMinConns  int32

	DBMaxConnIdleTime time.Duration
	DBMaxConnLifetime time.Duration
	DBHealthCheck     time.Duration

	Timezone *time.Location

	AuthMode        string
	APIKey          string
	SupabaseJWKSURL string
	SupabaseIssuer  string
	JWTSecret       string

	RunMigrations bool
	MigrationsDir string
}

func Load() Config {
	loadDotEnvFile(".env")

	// Zona bisnis WITA (Asia/Makassar, UTC+8)
	timezone, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		// Fallback to UTC if WITA timezone is not available
		timezone = time.UTC
	}

	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	issuer := os.Getenv("SUPABASE_ISSUER")
	if jwksURL == "" || issuer == "" {
		base := os.Getenv("SUPABASE_URL")
		if base == "" {
			base = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
		}
		if base != "" {
			if jwksURL == "" {
				jwksURL = strings.TrimRight(base, "/") + "/auth/v1/certs"
			}
			if issuer == "" {
				issuer = strings.TrimRight(base, "/") + "/auth/v1"
			}
		}
	}
	if jwksURL != "" {
		if u, err := url.Parse(jwksURL); err == nil {
			jwksURL = u.String()
		}
	}
	if issuer != "" {
		if u, err := url.Parse(issuer); err == nil {
			issuer = u.String()
		}
	}

	return Config{
		HTTPAddr: os.Getenv("HTTP_ADDR"),

		DatabaseURL: getString("DATABASE_URL", DefaultDatabaseURL),
		DBMaxConns:  int32(getInt("DB_MAX_CONNS", 10)),
		DBMinConns:  int32(getInt("DB_MIN_CONNS", 2)),

		DBMaxConnIdleTime: getDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		DBMaxConnLifetime: getDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		DBHealthCheck:     getDuration("DB_HEALTH_CHECK", 30*time.Second),

		Timezone: timezone,

		AuthMode:        getString("AUTH_MODE", "none"),
		APIKey:          os.Getenv("API_KEY"),
		SupabaseJWKSURL: jwksURL,
		SupabaseIssuer:  issuer,
		JWTSecret:       os.Getenv("JWT_SECRET"),

		RunMigrations: getBool("RUN_MIGRATIONS", false),
		MigrationsDir: getString("MIGRATIONS_DIR", "./migrations"),
	}
}

func loadDotEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := parseEnvLine(line)
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func parseEnvLine(line string) (string, string, bool) {
	i := strings.IndexByte(line, '=')
	if i <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:i])
	if key == "" {
		return "", "", false
	}
	val := strings.TrimSpace(line[i+1:])
	val = strings.TrimPrefix(val, `"`)
	val = strings.TrimSuffix(val, `"`)
	val = strings.TrimPrefix(val, `'`)
	val = strings.TrimSuffix(val, `'`)
	return key, val, true
}

func getString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
