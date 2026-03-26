package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"laundry-backend/internal/http/middleware"
	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/service"
)

type AuthHandler struct {
	svc       *service.AuthService
	jwtSecret string
}

func NewAuthHandler(svc *service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{svc: svc, jwtSecret: jwtSecret}
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		if strings.TrimSpace(h.jwtSecret) == "" {
			return httpapi.Internal("JWT_SECRET belum dikonfigurasi")
		}

		var body loginBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}

		row, err := h.svc.Login(r.Context(), body.Email, body.Password)
		if err != nil {
			return err
		}

		now := time.Now()
		claims := middleware.Claims{
			UserID: row.User.ID,
			Email:  row.User.Email,
			Role:   row.User.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			},
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
		if err != nil {
			return httpapi.Internal("Gagal membuat token sesi")
		}

		httpapi.WriteOK(w, http.StatusOK, map[string]any{
			"token": token,
			"user":  row.User,
		})
		return nil
	})
}

func (h *AuthHandler) Me() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		c, ok := middleware.GetClaims(r.Context())
		if !ok {
			return httpapi.Unauthorized("Unauthorized")
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]any{
			"id":    c.UserID,
			"email": c.Email,
			"role":  c.Role,
		})
		return nil
	})
}
