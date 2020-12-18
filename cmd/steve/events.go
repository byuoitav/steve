package main

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/steve"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func handleUpdates(m messenger, ds steve.DataService, log *zap.Logger) error {
	updates := make(chan steve.StateUpdate)
	g, ctx := errgroup.WithContext(context.Background())

	// read incoming updates
	g.Go(func() error {
		for {
			update, err := m.Next(ctx)
			if err != nil {
				return fmt.Errorf("unable to get next state update: %w", err)
			}

			updates <- update
		}
	})

	// publish events from updates
	errors := make(chan error)
	g.Go(func() error {
		for {
			select {
			case update := <-updates:
				go func() {
					if err := handleUpdate(m, ds, log, update); err != nil {
						select {
						case errors <- err:
						default:
						}
					}
				}()
			case err := <-errors:
				return err
			case <-ctx.Done():
				return nil
			}
		}
	})

	return g.Wait()
}

func handleUpdate(m messenger, ds steve.DataService, log *zap.Logger, update steve.StateUpdate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	states := make(map[string]bool, len(update.States))
	for _, state := range update.States {
		states[state] = true
	}

	configs, err := ds.EventConfigs(ctx, update.Room)
	if err != nil {
		return fmt.Errorf("unable to get event configs: %w", err)
	}

	sent := 0

	for _, config := range configs {
		if !statesMatch(states, config.MatchStates) {
			break
		}

		for _, event := range config.Events {
			event.Timestamp = update.Timestamp
			event.GeneratingSystem = update.GeneratingSystem

			if event.RoomID == "" {
				event.RoomID = update.Room
			}

			if err := m.Publish(ctx, event); err != nil {
				return fmt.Errorf("unable to publish event: %w", err)
			}

			sent++
		}
	}

	log.Info("Finished publishing events from update", zap.Int("sent", sent))
	return nil
}

func statesMatch(states map[string]bool, match []string) bool {
	for _, state := range match {
		if !states[state] {
			return false
		}
	}

	return true
}
