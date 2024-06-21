package rate

import (
	"fmt"
	"time"
)

type Rate struct {
	CurrencyFrom string
	CurrencyTo   string
	Rate         float32
	Time         time.Time
}

func (r Rate) String() string {
	return fmt.Sprintf("Rate<%s -> %s: %f>", r.CurrencyFrom, r.CurrencyTo, r.Rate)
}
