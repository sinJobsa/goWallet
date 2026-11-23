package storages

import "time"

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Balance struct {
	UserID   string
	Currency string
	Amount   float64
}

type TransactionType string

const (
	TxDeposit  TransactionType = "deposit"
	TxWithdraw TransactionType = "withdraw"
	TxExchange TransactionType = "exchange"
)

type Transaction struct {
	ID           string
	UserID       string
	Type         TransactionType
	FromCurrency string
	ToCurrency   string
	Amount       float64
	ResultAmount float64
	CreatedAt    time.Time
}
