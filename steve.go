package steve

import (
	"context"
	"time"
)

type EventConfig struct {
	MatchStates []string
	Events      []Event
}

type Event struct {
	Key      string
	Value    string
	DeviceID string
	Tags     []string

	// RoomID is passed through from the original event, but can be overridden in config
	RoomID string

	// GeneratingSystem is passed through from the original event
	GeneratingSystem string

	// Timestamp is passed through from the original event
	Timestamp time.Time
}

type DataService interface {
	EventConfigs(context.Context, string) ([]EventConfig, error)
}

type StateUpdate struct {
	GeneratingSystem string
	Timestamp        time.Time
	Room             string
	States           []string
}

type StateUpdateStreamer interface {
	Next(context.Context) (StateUpdate, error)
	Close() error
}

type EventPublisher interface {
	Publish(context.Context, Event) error
}
