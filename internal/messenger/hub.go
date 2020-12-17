package messenger

import (
	"context"
	"errors"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/steve"
)

type Messenger struct {
	GeneratingSystem string
	MatchKey         string

	m    *messenger.Messenger
	done chan struct{}
}

func New(hubURL string) (Messenger, error) {
	m, err := messenger.BuildMessenger(hubURL, base.Messenger, 512)
	if err != nil {
		return Messenger{}, err
	}

	m.SubscribeToRooms("*")

	return Messenger{
		m: m,
	}, nil
}

func (m *Messenger) Close() error {
	m.m.Kill()
	return nil
}

func (m *Messenger) Publish(ctx context.Context, event steve.Event) error {
	e := events.Event{
		GeneratingSystem: m.GeneratingSystem,
		Timestamp:        time.Now(),
		EventTags:        event.Tags,
		Key:              event.Key,
		Value:            event.Value,
		AffectedRoom:     events.GenerateBasicRoomInfo(event.RoomID),
		TargetDevice:     events.GenerateBasicDeviceInfo(event.DeviceID),
	}

	m.m.SendEvent(e)
	return nil
}

func (m *Messenger) Next(ctx context.Context) (steve.StateUpdate, error) {
	update := make(chan steve.StateUpdate)
	go func() {
		for {
			select {
			case <-m.done:
				return
			case <-ctx.Done():
				return
			default:
			}

			event := m.m.ReceiveEvent()
			if event.Key != m.MatchKey || event.AffectedRoom.RoomID == "" {
				continue
			}

			arr, ok := event.Data.([]interface{})
			if !ok {
				continue
			}

			var invalid bool
			var states []string
			for _, v := range arr {
				str, ok := v.(string)
				if !ok {
					invalid = true
					break
				}

				states = append(states, str)
			}

			if invalid {
				continue
			}

			update <- steve.StateUpdate{
				Room:   event.AffectedRoom.RoomID,
				States: states,
			}
			return
		}
	}()

	for {
		select {
		case update := <-update:
			return update, nil
		case <-m.done:
			return steve.StateUpdate{}, errors.New("messenger closed")
		case <-ctx.Done():
			return steve.StateUpdate{}, ctx.Err()
		}
	}
}
