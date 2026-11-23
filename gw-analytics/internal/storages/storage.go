package storages

import "context"

type Storage interface {
	SaveEvent(ctx context.Context, event *Event) error
	Close() error
}
