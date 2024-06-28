package fetchers_test

import (
	"context"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchRate_Failure(t *testing.T) {
	tests := []struct {
		name          string
		from          string
		to            string
		expectedError bool
	}{
		{
			name:          "invalid-currency",
			from:          "-",
			to:            "UAH",
			expectedError: true,
		},
		{
			name:          "unsupported-currency",
			from:          "USD",
			to:            "EUR",
			expectedError: true,
		},
	}

	b := fetchers.NewNBURateFetcher()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := b.FetchRate(context.Background(), tc.from, tc.to)
			if tc.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNBUFetchRate(t *testing.T) {
	nbu := fetchers.NewNBURateFetcher()
	rate, err := nbu.FetchRate(context.Background(), "USD", "UAH")
	require.NoError(t, err)
	assert.Greater(t, rate.Rate, float32(0))
}
