package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"laundry-backend/internal/httpapi"
	"laundry-backend/internal/repository"
	"laundry-backend/internal/service"
	"laundry-backend/internal/util"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) List() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		page := parseIntQuery(r, "page", 1)
		pageSize := parseIntQuery(r, "pageSize", 20)

		out, err := h.svc.List(r.Context(), q, page, pageSize)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *CustomerHandler) Get() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		out, err := h.svc.Get(r.Context(), id)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *CustomerHandler) RecentOrders() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		limit := parseIntQuery(r, "limit", 10)
		out, err := h.svc.RecentOrders(r.Context(), id, limit)
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

type customerBody struct {
	Name      string   `json:"name"`
	Phone     *string  `json:"phone"`
	Address   *string  `json:"address"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Email     *string  `json:"email"`
	Notes     *string  `json:"notes"`
}

func (h *CustomerHandler) Create() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var body customerBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Create(r.Context(), repositoryCustomerParams(body))
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusCreated, out)
		return nil
	})
}

func (h *CustomerHandler) Update() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		var body customerBody
		if err := decodeJSON(r, &body); err != nil {
			return err
		}
		out, err := h.svc.Update(r.Context(), id, repositoryCustomerParams(body))
		if err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, out)
		return nil
	})
}

func (h *CustomerHandler) Delete() http.Handler {
	return httpapi.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "id")
		if err := h.svc.Delete(r.Context(), id); err != nil {
			return err
		}
		httpapi.WriteOK(w, http.StatusOK, map[string]bool{"ok": true})
		return nil
	})
}

func repositoryCustomerParams(b customerBody) (p repository.CreateCustomerParams) {
	p.Name = b.Name
	p.Phone = trimPtr(b.Phone)
	p.Address = trimPtr(b.Address)
	p.Email = trimPtr(b.Email)
	p.Notes = trimPtr(b.Notes)
	p.Latitude = b.Latitude
	p.Longitude = b.Longitude
	return p
}

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return util.PtrString(s)
}
