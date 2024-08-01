package rate

import (
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"
)

const eventType = "RateFetched"

type rateData struct {
	From string    `json:"from"`
	To   string    `json:"to"`
	Rate float32   `json:"body"`
	Time time.Time `json:"time"`
}

type rateFetchedEvent struct {
	broker.Event
	Data rateData `json:"data"`
}
