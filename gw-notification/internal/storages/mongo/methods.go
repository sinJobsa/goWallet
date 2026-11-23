package mongo

import (
	"context"
	"fmt"

	"github.com/sinJobsa/goWallet/gw-notification/internal/storages"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Storage) SaveTransfer(ctx context.Context, transfer *storages.Transfer) error {
	filter := bson.M{"_id": transfer.TransactionID}
	opts := options.Replace().SetUpsert(true)

	_, err := s.collection.ReplaceOne(ctx, filter, transfer, opts)
	if err != nil {
		return fmt.Errorf("replace transfer %s: %w", transfer.TransactionID, err)
	}
	return nil
}
