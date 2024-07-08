package user

import "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"

const (
	subscribedEventType   = "Subscribe"
	unsubscribedEventType = "Unsubscribe"
)

type subscriberData struct {
	Email string `json:"email"`
}

type subscribeEvent struct {
	broker.Event
	Data subscriberData `json:"data"`
}
