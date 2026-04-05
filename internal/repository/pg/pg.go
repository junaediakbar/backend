package pg

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

// timestampAsUTCWall mengirim instant sebagai TIMESTAMP naif dengan jam dinding UTC.
// pgx Timestamp mengabaikan lokasi (discardTimeZone); tanpa .UTC() batas filter WITA
// salah ter-encode dan tidak selaras dengan created_at dari PostgreSQL now() (Neon).
func timestampAsUTCWall(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}

func New(pool *pgxpool.Pool) *DB {
	return &DB{Pool: pool}
}
