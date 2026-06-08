package deliveryservice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"express_1", CategoryExpress1, true},
		{"EXPRESS_2", CategoryExpress2, true},
		{" cepat ", CategoryCepat, true},
		{"reguler", CategoryReguler, true},
		{"", "", false},
		{"invalid", "", false},
	}
	for _, tc := range tests {
		got, ok := NormalizeCategory(tc.in)
		require.Equal(t, tc.ok, ok, tc.in)
		if ok {
			require.Equal(t, tc.want, got)
		}
	}
}

func TestEstimateDaysFor(t *testing.T) {
	days, ok := EstimateDaysFor(CategoryCepat)
	require.True(t, ok)
	require.Equal(t, 5, days)

	days, ok = EstimateDaysFor(CategoryReguler)
	require.True(t, ok)
	require.Equal(t, 7, days)

	_, ok = EstimateDaysFor("unknown")
	require.False(t, ok)
}

func TestSurchargePercentFor(t *testing.T) {
	pct, ok := SurchargePercentFor(CategoryCepat)
	require.True(t, ok)
	require.Equal(t, 10, pct)

	pct, ok = SurchargePercentFor(CategoryExpress1)
	require.True(t, ok)
	require.Equal(t, 100, pct)

	pct, ok = SurchargePercentFor(CategoryReguler)
	require.True(t, ok)
	require.Equal(t, 0, pct)
}

func TestApplySurcharge(t *testing.T) {
	require.Equal(t, "10000.00", ApplySurcharge("10000.00", 0))
	require.Equal(t, "11000.00", ApplySurcharge("10000.00", 10))
	require.Equal(t, "12500.00", ApplySurcharge("10000.00", 25))
	require.Equal(t, "15000.00", ApplySurcharge("10000.00", 50))
	require.Equal(t, "20000.00", ApplySurcharge("10000.00", 100))
}
