package broker

type Command struct {
	ID        string `json:"commandID"`
	Type      string `json:"commandType"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}

type Event struct {
	ID        string `json:"eventID"`
	Type      string `json:"eventType"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}
