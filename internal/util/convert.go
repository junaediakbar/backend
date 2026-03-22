package util

import (
	"fmt"
	"math"
)

func Money2(v float64) string {
	if !isFinite(v) {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", v)
}

func Float6(v float64) string {
	if !isFinite(v) {
		return "0.000000"
	}
	return fmt.Sprintf("%.6f", v)
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func PtrString(s string) *string { return &s }

func PtrFloat64(v float64) *float64 { return &v }

func PtrInt(v int) *int { return &v }
