package messenger

import (
	"context"
	"errors"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/steve"
	"go.uber.org/zap"
)

type Messenger struct {
	MatchKey string
	Log      *zap.Logger

	m    *messenger.Messenger
	done chan struct{}
}

func New(hubURL string) (*Messenger, error) {
	m, err := messenger.BuildMessenger(hubURL, base.Messenger, 512)
	if err != nil {
		return nil, err
	}

	m.SubscribeToRooms("*")

	return &Messenger{
		m: m,
	}, nil
}

func (m *Messenger) Close() error {
	m.m.UnsubscribeFromRooms("*")
	m.m.Kill()
	return nil
}

func (m *Messenger) Publish(ctx context.Context, event steve.Event) error {
	e := events.Event{
		Timestamp:        event.Timestamp,
		GeneratingSystem: event.GeneratingSystem,
		EventTags:        event.Tags,
		Key:              event.Key,
		Value:            event.Value,
		AffectedRoom:     events.GenerateBasicRoomInfo(event.RoomID),
		TargetDevice:     events.GenerateBasicDeviceInfo(event.DeviceID),
	}

	m.Log.Debug("Publishing", zap.Any("event", e))

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

			m.Log.Debug("Received state update", zap.String("room", event.AffectedRoom.RoomID), zap.Strings("states", states))

			update <- steve.StateUpdate{
				GeneratingSystem: event.GeneratingSystem,
				Timestamp:        event.Timestamp,
				Room:             event.AffectedRoom.RoomID,
				States:           states,
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
