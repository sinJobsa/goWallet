package storages

import (
	"context"
	"errors"
)

var ErrAlreadyExists = errors.New("transfer already exists")

type Storage interface {
	SaveTransfer(ctx context.Context, transfer *Transfer) error
	Close(ctx context.Context) error
}
