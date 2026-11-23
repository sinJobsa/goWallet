package storages

import (
	"context"
	"errors"
)

var (
	ErrNotFound            = errors.New("record not found")
	ErrAlreadyExists       = errors.New("record already exists")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Storage interface {
	//Users
	CreateUser(ctx context.Context, user *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)

	//Balances
	GetBalances(ctx context.Context, userID string) (map[string]float64, error)
	GetBalance(ctx context.Context, userID, currency string) (float64, error)

	//Operations
	Deposit(ctx context.Context, userID, currency string, amount float64) (map[string]float64, error)
	Withdraw(ctx context.Context, userID, currency string, amount float64) (map[string]float64, error)
	Exchange(ctx context.Context, userID, from, to string, amount, rate float64) (newBalances map[string]float64, resultAmount float64, err error)

	//Transactions
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
}
