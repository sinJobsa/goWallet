package storages

import "time"

type Event struct {
	TransactionID string    `ch:"transaction_id"`
	UserID        string    `ch:"user_id"`
	FromCurrency  string    `ch:"from_currency"`
	ToCurrency    string    `ch:"to_currency"`
	Amount        float64   `ch:"amount"`
	Status        string    `ch:"status"`
	CreatedAt     time.Time `ch:"created_at"`
	ReceivedAt    time.Time `ch:"received_at"`
	LatencyMs     int64     `ch:"latency_ms"`
}
