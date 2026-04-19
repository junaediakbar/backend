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

		emp := row.Employee
		now := time.Now()
		claims := middleware.Claims{
			UserID:     emp.ID,
			Email:      emp.Email,
			Role:       emp.Role,
			EmployeeID: emp.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			},
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
		if err != nil {
			return httpapi.Internal("Gagal membuat token sesi")
		}

		userOut := map[string]any{
			"id":       emp.ID,
			"name":     emp.Name,
			"email":    emp.Email,
			"role":     emp.Role,
			"isActive": emp.IsActive,
		}
		if emp.Role == "employee" {
			userOut["employeeId"] = emp.ID
		}

		httpapi.WriteOK(w, http.StatusOK, map[string]any{
			"token": token,
			"user":  userOut,
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
		out := map[string]any{
			"id":    c.UserID,
			"email": c.Email,
			"role":  c.Role,
		}
		if c.EmployeeID != "" {
			out["employeeId"] = c.EmployeeID
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}
