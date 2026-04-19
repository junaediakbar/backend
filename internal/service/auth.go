package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/repository"
)

type AuthService struct {
	employees repository.EmployeeRepository
}

func NewAuthService(employees repository.EmployeeRepository) *AuthService {
	return &AuthService{employees: employees}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*repository.EmployeeAuthRow, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, httpapi.BadRequest("validation_error", "Email wajib diisi", nil)
	}
	if strings.TrimSpace(password) == "" {
		return nil, httpapi.BadRequest("validation_error", "Password wajib diisi", nil)
	}

	row, err := s.employees.GetByEmailForAuth(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.Unauthorized("Email atau password salah")
		}
		return nil, err
	}
	if !row.Employee.IsActive {
		return nil, httpapi.Forbidden("Akun nonaktif")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return nil, httpapi.Unauthorized("Email atau password salah")
	}
	return row, nil
}
