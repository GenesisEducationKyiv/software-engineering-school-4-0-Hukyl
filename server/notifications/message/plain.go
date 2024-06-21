package message

import (
	"fmt"

	"github.com/Hukyl/genesis-kma-school-entry/models"
)

// PlainRateMessage is a message that contains the exchange rate between two currencies.
type PlainRateMessage struct {
	rate *models.Rate
}

func (m *PlainRateMessage) SetRate(rate *models.Rate) {
	m.rate = rate
}

func (m *PlainRateMessage) Subject() string {
	return fmt.Sprintf("%s-%s exchange rate", m.rate.CurrencyFrom, m.rate.CurrencyTo)
}

func (m *PlainRateMessage) String() string {
	return fmt.Sprintf(
		"1 %s = %f %s",
		m.rate.CurrencyFrom,
		m.rate.Rate,
		m.rate.CurrencyTo,
	)
}
