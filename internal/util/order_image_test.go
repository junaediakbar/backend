package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeOrderImageColumn(t *testing.T) {
	t.Parallel()

	s := "https://example.com/a.jpg"
	first, all := NormalizeOrderImageColumn(&s)
	require.NotNil(t, first)
	require.Equal(t, s, *first)
	require.Equal(t, []string{s}, all)

	legacy := `["https://x/a","https://x/b"]`
	first, all = NormalizeOrderImageColumn(&legacy)
	require.NotNil(t, first)
	require.Equal(t, "https://x/a", *first)
	require.Equal(t, []string{"https://x/a", "https://x/b"}, all)

	empty := ""
	first, all = NormalizeOrderImageColumn(&empty)
	require.Nil(t, first)
	require.Nil(t, all)
}

func TestEncodeOrderImagesJSON(t *testing.T) {
	t.Parallel()

	require.Nil(t, EncodeOrderImagesJSON(nil))
	require.Nil(t, EncodeOrderImagesJSON([]string{}))

	out := EncodeOrderImagesJSON([]string{" https://a ", "https://b"})
	require.NotNil(t, out)
	require.Equal(t, `["https://a","https://b"]`, *out)
}
