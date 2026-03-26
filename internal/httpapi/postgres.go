package httpapi

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// MapPostgresError mengubah error Postgres menjadi AppError agar klien tidak hanya melihat "internal".
// Mengembalikan nil jika bukan error Postgres yang dikenali (biarkan handler log stack).
func MapPostgresError(err error) *AppError {
	if err == nil {
		return nil
	}
	var pe *pgconn.PgError
	if !errors.As(err, &pe) {
		return nil
	}
	switch pe.Code {
	case "42P01": // undefined_table
		return Internal("Tabel basis data tidak ditemukan. Jalankan migrasi backend (migrate).")
	case "42703": // undefined_column
		return Internal("Kolom basis data tidak ditemukan. Jalankan migrasi backend (migrate).")
	case "3D000": // invalid_catalog_name
		return Internal("Basis data tidak ditemukan. Periksa DATABASE_URL.")
	case "3F000": // invalid_schema_name
		return Internal("Skema laundry_backend tidak ada. Jalankan migrasi backend (migrate).")
	case "23503": // foreign_key_violation
		return BadRequest("constraint", "Data tidak valid: referensi ke tabel lain tidak ditemukan.", map[string]string{"detail": pe.Message})
	case "23505": // unique_violation
		return Conflict("Data bentrok dengan rekaman yang sudah ada.")
	case "23514": // check_violation
		return BadRequest("constraint", "Data tidak memenuhi aturan basis data.", map[string]string{"detail": pe.Message})
	case "23502": // not_null_violation
		return BadRequest("constraint", "Data wajib tidak boleh kosong (basis data).", map[string]string{"detail": pe.Message})
	case "42804": // datatype_mismatch
		return Internal("Tipe data di basis data tidak cocok. Jalankan migrasi backend atau hubungi admin.")
	case "22P02": // invalid_text_representation
		return BadRequest("constraint", "Format nilai tidak valid untuk basis data.", map[string]string{"detail": pe.Message})
	default:
		return nil
	}
}

// MapCommonDriverError menangkap pesan koneksi umum (bukan *pgconn.PgError).
func MapCommonDriverError(err error) *AppError {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "dial tcp") && strings.Contains(msg, "connect: connection refused") {
		return Internal("Tidak dapat terhubung ke basis data. Pastikan PostgreSQL berjalan dan DATABASE_URL benar.")
	}
	if strings.Contains(msg, "password authentication failed") {
		return Internal("Autentikasi ke basis data gagal. Periksa user/password di DATABASE_URL.")
	}
	return nil
}
