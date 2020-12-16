package couch

import (
	"context"
	"fmt"

	"github.com/byuoitav/steve"
	"github.com/go-kivik/kivik/v3"
)

type DataService struct {
	client   *kivik.Client
	database string
}

func New(ctx context.Context, url string, opts ...Option) (*DataService, error) {
	client, err := kivik.New("couch", url)
	if err != nil {
		return nil, fmt.Errorf("unable to build client: %w", err)
	}

	return NewWithClient(ctx, client, opts...)
}

func NewWithClient(ctx context.Context, client *kivik.Client, opts ...Option) (*DataService, error) {
	options := options{
		database: "steve",
	}

	for _, o := range opts {
		o.apply(&options)
	}

	if options.authFunc != nil {
		if err := client.Authenticate(ctx, options.authFunc); err != nil {
			return nil, fmt.Errorf("unable to authenticate: %w", err)
		}
	}

	return &DataService{
		client:   client,
		database: options.database,
	}, nil
}

func (d *DataService) EventConfigs(ctx context.Context, room string) ([]steve.EventConfig, error) {
	var doc doc

	db := d.client.DB(ctx, d.database)
	if err := db.Get(ctx, room).ScanDoc(&doc); err != nil {
		return nil, fmt.Errorf("unable to get/scan room: %w", err)
	}

	res := make([]steve.EventConfig, len(doc.Events))
	for i, config := range doc.Events {
		res[i].MatchStates = config.MatchStates

		res[i].Events = make([]steve.Event, len(config.Events))
		for j, event := range config.Events {
			res[i].Events[j] = steve.Event{
				Key:   event.Key,
				Value: event.Value,
			}
		}
	}

	return res, nil
}

type doc struct {
	ID     string        `json:"_id"`
	Events []eventConfig `json:"events"`
}

type eventConfig struct {
	MatchStates []string `json:"matchStates"`
	Events      []event  `json:"events"`
}

type event struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
