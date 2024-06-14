package notifications

import (
	"fmt"

	"github.com/Hukyl/genesis-kma-school-entry/rate"
)

// RateMessage is a message that contains the exchange rate between two currencies.
type RateMessage struct {
	rate rate.Rate
}

func (m RateMessage) Rate() rate.Rate {
	return m.rate
}

func (m RateMessage) String() string {
	return fmt.Sprintf(
		"1 %s = %f %s",
		m.rate.CurrencyFrom,
		m.rate.Rate,
		m.rate.CurrencyTo,
	)
}

func NewRateMessage(rate rate.Rate) RateMessage {
	return RateMessage{rate: rate}
}
