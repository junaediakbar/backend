package pg

import "github.com/jackc/pgx/v5/pgxpool"

type DB struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *DB {
	return &DB{Pool: pool}
}
