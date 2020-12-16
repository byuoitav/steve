package steve

import (
	"context"
)

type EventConfig struct {
	MatchStates []string
	Events      []Event
}

type Event struct {
	Key   string
	Value string
}

type DataService interface {
	EventConfigs(context.Context, string) ([]EventConfig, error)
}

type StateUpdate struct {
	Room   string
	States []string
}

type StateUpdateStreamer interface {
	Next(context.Context) (StateUpdate, error)
}

type EventPublisher interface {
	Publish(context.Context, Event) error
}
