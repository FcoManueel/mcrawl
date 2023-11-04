package urlparse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedHost string
	}{
		{
			"good-input",
			"http://google.com",
			"google.com",
		},
		{
			"no-schema",
			"google.com",
			"google.com",
		},
		{
			"www",
			"  www.google.com  ",
			"google.com",
		},
		{
			"uppercase",
			"GOOGLE.COM",
			"google.com",
		},
		{
			"https",
			"https://google.com",
			"google.com",
		},
		{
			"trailing-slash",
			"google.com/",
			"google.com",
		},
		{
			"spaces",
			"  www.google.com  ",
			"google.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Normalize(tt.input)
			require.NoError(t, err)
			require.Equal(t, got.Host, tt.expectedHost)
		})
	}
}
