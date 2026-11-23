package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/storages"
)

func (s *Storage) CreateUser(ctx context.Context, user *storages.User) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "INSERT INTO users (id, username, email, password_hash, created_at) VALUES ($1, $2,$3,$4,$5)",
		user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storages.ErrAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO balances (user_id, currency, amount) VALUES
         ($1, 'USD', 0), ($1, 'RUB', 0), ($1, 'EUR', 0)`,
		user.ID)
	if err != nil {
		return fmt.Errorf("insert balances: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil

}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*storages.User, error) {
	var u storages.User
	err := s.pool.QueryRow(ctx, "SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1", username).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storages.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id string) (*storages.User, error) {
	var u storages.User
	err := s.pool.QueryRow(ctx, "SELECT id, username, password_hash, created_at FROM users WHERE id = $1", id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storages.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (s *Storage) GetBalances(ctx context.Context, userID string) (map[string]float64, error) {
	rows, err := s.pool.Query(ctx, "SELECT currency, amount FROM balances WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("get user balances: %w", err)
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var balance float64
		var currency string
		if err := rows.Scan(&currency, &balance); err != nil {
			return nil, fmt.Errorf("scan balances: %w", err)
		}
		result[currency] = balance
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get user balances: %w", err)
	}
	return result, nil
}

func (s *Storage) GetBalance(ctx context.Context, userID, currency string) (float64, error) {
	var amount float64
	err := s.pool.QueryRow(ctx, "SELECT amount FROM balances WHERE user_id = $1 AND currency = $2", userID, currency).Scan(&amount)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, storages.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get user balance: %w", err)
	}
	return amount, nil
}

func (s *Storage) Deposit(ctx context.Context, userID, currency string, amount float64) (map[string]float64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, "UPDATE balances SET amount = amount + $1 WHERE user_id = $2 AND currency = $3", amount, userID, currency)
	if err != nil {
		return nil, fmt.Errorf("update user balances: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, storages.ErrNotFound
	}

	txID := uuid.NewString()
	_, err = tx.Exec(ctx, "INSERT INTO transactions (id, user_id, type, from_currency, to_currency, amount, result_amount, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		txID, userID, storages.TxDeposit, currency, currency, amount, amount)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	balances, err := s.getBalancesTx(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user balances: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return balances, nil

}

func (s *Storage) getBalancesTx(ctx context.Context, tx pgx.Tx, userID string) (map[string]float64, error) {
	rows, err := tx.Query(ctx, "SELECT currency, amount FROM balances WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("get user balances: %w", err)
	}
	defer rows.Close()
	result := make(map[string]float64)
	for rows.Next() {
		var amount float64
		var currency string
		if err := rows.Scan(&currency, &amount); err != nil {
			return nil, fmt.Errorf("scan balances: %w", err)
		}
		result[currency] = amount
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get user balances: %w", err)
	}
	return result, nil
}

func (s *Storage) Withdraw(ctx context.Context, userID, currency string, amount float64) (map[string]float64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, "UPDATE balances SET amount = amount - $1 WHERE user_id = $2 AND currency = $3 AND amount >= $1", amount, userID, currency)
	if err != nil {
		return nil, fmt.Errorf("update user balances: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, storages.ErrInsufficientBalance
	}

	txID := uuid.NewString()
	_, err = tx.Exec(ctx, "INSERT INTO transactions (id, user_id, type, from_currency, to_currency, amount, result_amount, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		txID, userID, storages.TxWithdraw, currency, currency, amount, amount)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	balances, err := s.getBalancesTx(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user balances: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return balances, nil
}

func (s *Storage) Exchange(ctx context.Context, userID, from, to string, amount, rate float64) (newBalances map[string]float64, resultAmount float64, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, "UPDATE balances SET amount = amount - $1 WHERE user_id = $2 AND currency = $3 AND amount >= $1", amount, userID, from)
	if err != nil {
		return nil, 0, fmt.Errorf("update user balances: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, 0, storages.ErrInsufficientBalance
	}

	resultAmount = amount * rate

	tag, err = tx.Exec(ctx, "INSERT INTO balances (user_id, currency, amount) VALUES ($1, $2, $3) ON CONFLICT (user_id, currency) DO UPDATE SET amount = balances.amount + EXCLUDED.amount", userID, to, resultAmount)
	if err != nil {
		return nil, 0, fmt.Errorf("update user balances: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, 0, fmt.Errorf("balance %s not found", to)
	}

	txID := uuid.NewString()
	_, err = tx.Exec(ctx, "INSERT INTO transactions (id, user_id, type, from_currency, to_currency, amount, result_amount, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())",
		txID, userID, storages.TxExchange, from, to, amount, resultAmount)
	if err != nil {
		return nil, 0, fmt.Errorf("insert transaction: %w", err)
	}

	newBalances, err = s.getBalancesTx(ctx, tx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("get user balances: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("commit tx: %w", err)
	}
	return newBalances, resultAmount, nil
}

func (s *Storage) GetTransaction(ctx context.Context, id string) (*storages.Transaction, error) {
	var t storages.Transaction
	err := s.pool.QueryRow(ctx, "SELECT id, user_id, type, from_currency, to_currency, amount, result_amount, created_at FROM transactions WHERE id = $1", id).
		Scan(&t.ID, &t.UserID, &t.Type, &t.FromCurrency, &t.ToCurrency, &t.Amount, &t.ResultAmount, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storages.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	return &t, nil
}
