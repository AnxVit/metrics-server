package audit

import "time"

type Publisher interface {
	Register(Audit)
	Notify(Event)
	Wait()
}

type Audit interface {
	GetID() string
	Update(event Event)
}

type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

func MakeEvent(metrics []string, ipAddress string) Event {
	return Event{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddress,
	}
}
