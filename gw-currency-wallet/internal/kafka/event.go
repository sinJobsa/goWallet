package kafka

import "time"

type Event struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	FromCurrency  string    `json:"from_currency"`
	ToCurrency    string    `json:"to_currency"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
}

const (
	StatusSuccess = "success"
)
