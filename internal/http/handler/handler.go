package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"laundry-backend/internal/httpapi"
)

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return httpapi.BadRequest("invalid_json", "Body JSON tidak valid", nil)
	}
	return nil
}

func parseIntQuery(r *http.Request, key string, def int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func parseBoolQuery(r *http.Request, key string) (*bool, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, httpapi.BadRequest("validation_error", "Query param tidak valid", map[string]string{"param": key})
	}
	return &v, nil
}

// parseDateQuery menginterpretasikan YYYY-MM-DD sebagai hari kalender di lok (mis. WITA),
// lalu mengembalikan momen waktu yang benar untuk dibandingkan dengan created_at (UTC wall clock di DB).
func parseDateQuery(r *http.Request, key string, endOfDay bool, loc *time.Location) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	if loc == nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation("2006-01-02", raw, loc)
	if err != nil {
		return nil, httpapi.BadRequest("validation_error", "Tanggal tidak valid", map[string]string{"param": key})
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return &t, nil
}
