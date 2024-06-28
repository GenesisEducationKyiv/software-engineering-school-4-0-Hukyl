package fetchers_test

import (
	"context"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
	"github.com/stretchr/testify/assert"
)

func TestFetchRate(t *testing.T) {
	tests := []struct {
		name          string
		from          string
		to            string
		expectedError bool
	}{
		{
			name:          "unsupported-currency",
			from:          "-",
			to:            "UAH",
			expectedError: true,
		},
		{
			name:          "no-API-key",
			from:          "USD",
			to:            "EUR",
			expectedError: true,
		},
	}

	b := fetchers.NewCurrencyBeaconFetcher("")

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
