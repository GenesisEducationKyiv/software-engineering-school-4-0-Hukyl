package models

import "fmt"

type Rate struct {
	ID           uint   `gorm:"primaryKey"`
	CurrencyFrom string `gorm:"column:cc_from"`
	CurrencyTo   string `gorm:"column:cc_to"`
	Rate         float32
	Created      int64 `gorm:"autoCreateTime"` // Use unix seconds as creating time
}

func (r Rate) String() string {
	return fmt.Sprintf("Rate<%d, %s-%s: %f>", r.ID, r.CurrencyFrom, r.CurrencyTo, r.Rate)
}
