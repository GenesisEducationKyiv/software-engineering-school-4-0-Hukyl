package message

import (
	"fmt"

	"github.com/Hukyl/genesis-kma-school-entry/internal/models"
)

// PlainRate is a message that contains the exchange rate between two currencies.
type PlainRate struct {
	rate *models.Rate
}

func (m *PlainRate) SetRate(rate *models.Rate) {
	m.rate = rate
}

func (m *PlainRate) Subject() string {
	return fmt.Sprintf("%s-%s exchange rate", m.rate.CurrencyFrom, m.rate.CurrencyTo)
}

func (m *PlainRate) String() string {
	return fmt.Sprintf(
		"1 %s = %f %s",
		m.rate.CurrencyFrom,
		m.rate.Rate,
		m.rate.CurrencyTo,
	)
}
