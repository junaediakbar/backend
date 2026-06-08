package deliveryservice

import (
	"fmt"
	"strings"
)

const (
	CategoryExpress1 = "express_1"
	CategoryExpress2 = "express_2"
	CategoryExpress3 = "express_3"
	CategoryCepat    = "cepat"
	CategoryReguler  = "reguler"
)

type Category struct {
	Code             string
	Label            string
	EstimateDays     int
	SurchargePercent int
}

var All = []Category{
	{Code: CategoryExpress1, Label: "Express 1", EstimateDays: 1, SurchargePercent: 100},
	{Code: CategoryExpress2, Label: "Express 2", EstimateDays: 2, SurchargePercent: 50},
	{Code: CategoryExpress3, Label: "Express 3", EstimateDays: 3, SurchargePercent: 25},
	{Code: CategoryCepat, Label: "Cepat", EstimateDays: 5, SurchargePercent: 10},
	{Code: CategoryReguler, Label: "Reguler", EstimateDays: 7, SurchargePercent: 0},
}

func NormalizeCategory(code string) (string, bool) {
	code = strings.TrimSpace(strings.ToLower(code))
	for _, c := range All {
		if c.Code == code {
			return c.Code, true
		}
	}
	return "", false
}

func EstimateDaysFor(code string) (int, bool) {
	code, ok := NormalizeCategory(code)
	if !ok {
		return 0, false
	}
	for _, c := range All {
		if c.Code == code {
			return c.EstimateDays, true
		}
	}
	return 0, false
}

func LabelFor(code string) string {
	code, ok := NormalizeCategory(code)
	if !ok {
		return ""
	}
	for _, c := range All {
		if c.Code == code {
			return c.Label
		}
	}
	return ""
}

func SurchargePercentFor(code string) (int, bool) {
	code, ok := NormalizeCategory(code)
	if !ok {
		return 0, false
	}
	for _, c := range All {
		if c.Code == code {
			return c.SurchargePercent, true
		}
	}
	return 0, false
}

// ApplySurcharge menambah persentase markup ke subtotal (string uang 2 desimal).
func ApplySurcharge(subtotal string, percent int) string {
	if percent <= 0 {
		return subtotal
	}
	cents := parseMoneyCents(subtotal)
	if cents <= 0 {
		return subtotal
	}
	// Pembulatan setengah ke atas per sen.
	totalCents := (cents*int64(100+percent) + 50) / 100
	return formatMoneyCents(totalCents)
}

func parseMoneyCents(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var whole int64
	var frac int64
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	for _, ch := range parts[0] {
		if ch < '0' || ch > '9' {
			continue
		}
		whole = whole*10 + int64(ch-'0')
	}
	if len(parts) == 2 {
		fs := parts[1]
		if len(fs) > 2 {
			fs = fs[:2]
		}
		for len(fs) < 2 {
			fs += "0"
		}
		for _, ch := range fs {
			if ch < '0' || ch > '9' {
				continue
			}
			frac = frac*10 + int64(ch-'0')
		}
	}
	out := whole*100 + frac
	if neg {
		return -out
	}
	return out
}

func formatMoneyCents(c int64) string {
	neg := c < 0
	if neg {
		c = -c
	}
	whole := c / 100
	frac := c % 100
	if neg {
		return fmt.Sprintf("-%d.%02d", whole, frac)
	}
	return fmt.Sprintf("%d.%02d", whole, frac)
}
