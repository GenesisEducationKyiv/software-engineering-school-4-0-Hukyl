package user

import (
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker"
)

const subTimeout = 5 * time.Second

const (
	subscribedEventType             = "Subscribe"
	unsubscribedEventType           = "Unsubscribe"
	compensateSubscribedEventType   = "CompensateSubscribe"
	compensateUnsubscribedEventType = "CompensateUnsubscribe"
)

type subscriberData struct {
	Email string `json:"email"`
}

type subscribeEvent struct {
	broker.Event
	Data subscriberData `json:"data"`
}
