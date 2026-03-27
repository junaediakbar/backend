package util

import (
	"encoding/json"
	"strings"
)

// NormalizeOrderImageColumn reads the `orders.image` column: legacy single URL,
// or JSON array of URLs (max 3 stored by the API).
func NormalizeOrderImageColumn(raw *string) (first *string, all []string) {
	if raw == nil {
		return nil, nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil, nil
	}
	if strings.HasPrefix(s, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(s), &arr); err != nil {
			return nil, nil
		}
		out := make([]string, 0, len(arr))
		for _, u := range arr {
			u = strings.TrimSpace(u)
			if u != "" {
				out = append(out, u)
			}
		}
		if len(out) == 0 {
			return nil, nil
		}
		f := out[0]
		return &f, out
	}
	return &s, []string{s}
}

// EncodeOrderImagesJSON stores URLs as a JSON array string in `orders.image`.
func EncodeOrderImagesJSON(urls []string) *string {
	clean := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u != "" {
			clean = append(clean, u)
		}
	}
	if len(clean) == 0 {
		return nil
	}
	b, err := json.Marshal(clean)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
