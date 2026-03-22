package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"laundry-backend/internal/httpapi"
)

type AuthConfig struct {
	Mode           string
	APIKey         string
	SupabaseJWKS   string
	SupabaseIssuer string
	JWTSecret      string
}

type ctxKey string

const authClaimsKey ctxKey = "authClaims"

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type JWKSProvider struct {
	once    sync.Once
	jwks    keyfunc.Keyfunc
	jwksErr error
}

func (p *JWKSProvider) Get(url string) (keyfunc.Keyfunc, error) {
	p.once.Do(func() {
		k, err := keyfunc.NewDefault([]string{url})
		if err != nil {
			p.jwksErr = err
			return
		}
		p.jwks = k
	})
	return p.jwks, p.jwksErr
}

func WithAuth(cfg AuthConfig, jwksProvider *JWKSProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
			case "jwt":
				if cfg.JWTSecret == "" {
					httpapi.WriteError(w, http.StatusInternalServerError, "auth_config", "JWT_SECRET belum dikonfigurasi", nil)
					return
				}
				claims, err := validateHMACJWT(r, cfg.JWTSecret)
				if err != nil {
					httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
					return
				}
				ctx := context.WithValue(r.Context(), authClaimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			case "", "supabase_jwks":
				if cfg.SupabaseJWKS == "" {
					httpapi.WriteError(w, http.StatusInternalServerError, "auth_config", "SUPABASE_JWKS_URL belum dikonfigurasi", nil)
					return
				}
				claims, err := validateJWT(r, cfg, jwksProvider)
				if err != nil {
					httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
					return
				}
				ctx := context.WithValue(r.Context(), authClaimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			case "api_key":
				if cfg.APIKey == "" {
					httpapi.WriteError(w, http.StatusInternalServerError, "auth_config", "API_KEY belum dikonfigurasi", nil)
					return
				}
				if r.Header.Get("X-API-Key") != cfg.APIKey {
					httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
					return
				}
				next.ServeHTTP(w, r)
				return
			case "none":
				next.ServeHTTP(w, r)
				return
			default:
				httpapi.WriteError(w, http.StatusInternalServerError, "auth_config", "AUTH_MODE tidak dikenali", nil)
				return
			}
		})
	}
}

func GetClaims(ctx context.Context) (*Claims, bool) {
	v := ctx.Value(authClaimsKey)
	c, ok := v.(*Claims)
	return c, ok
}

func validateJWT(r *http.Request, cfg AuthConfig, jwksProvider *JWKSProvider) (*Claims, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return nil, errors.New("missing auth header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid auth header")
	}
	raw := strings.TrimSpace(parts[1])
	if raw == "" {
		return nil, errors.New("missing token")
	}

	jwks, err := jwksProvider.Get(cfg.SupabaseJWKS)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(raw, &Claims{}, jwks.Keyfunc, jwt.WithLeeway(30*time.Second))
	if err != nil {
		return nil, err
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if cfg.SupabaseIssuer != "" && c.Issuer != cfg.SupabaseIssuer {
		return nil, errors.New("invalid issuer")
	}
	return c, nil
}

func validateHMACJWT(r *http.Request, secret string) (*Claims, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return nil, errors.New("missing auth header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid auth header")
	}
	raw := strings.TrimSpace(parts[1])
	if raw == "" {
		return nil, errors.New("missing token")
	}

	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithLeeway(30*time.Second))
	if err != nil {
		return nil, err
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if c.UserID == "" || c.Email == "" || c.Role == "" {
		return nil, errors.New("missing claims")
	}
	return c, nil
}
