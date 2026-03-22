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
	users repository.UserRepository
}

func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

type LoginResult struct {
	User  repository.UserAuthRow
	Token string
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*repository.UserAuthRow, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, httpapi.BadRequest("validation_error", "Email wajib diisi", nil)
	}
	if strings.TrimSpace(password) == "" {
		return nil, httpapi.BadRequest("validation_error", "Password wajib diisi", nil)
	}

	row, err := s.users.GetByEmailForAuth(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpapi.Unauthorized("Email atau password salah")
		}
		return nil, err
	}
	if !row.User.IsActive {
		return nil, httpapi.Forbidden("User nonaktif")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return nil, httpapi.Unauthorized("Email atau password salah")
	}
	return row, nil
}
