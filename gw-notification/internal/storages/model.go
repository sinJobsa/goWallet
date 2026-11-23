package storages

import "time"

type Transfer struct {
	TransactionID string    `bson:"_id"`
	UserID        string    `bson:"user_id"`
	FromCurrency  string    `bson:"from_currency"`
	ToCurrency    string    `bson:"to_currency"`
	Amount        float64   `bson:"amount"`
	Timestamp     time.Time `bson:"timestamp"`
	Status        string    `bson:"status"`
	ReceivedAt    time.Time `bson:"received_at"`
}
