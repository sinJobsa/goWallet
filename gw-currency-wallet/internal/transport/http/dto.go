package http

// тело POST /api/v1/register
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

// тело POST /api/v1/login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// тело POST /api/v1/wallet/deposit
type DepositRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required,oneof=USD RUB EUR"`
}

// тело POST /api/v1/wallet/withdraw
type WithdrawRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required,oneof=USD RUB EUR"`
}

// тело POST /api/v1/exchange
type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" binding:"required,oneof=USD RUB EUR"`
	ToCurrency   string  `json:"to_currency" binding:"required,oneof=USD RUB EUR"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
}

// простой ответ с сообщением
type MessageResponse struct {
	Message string `json:"message"`
}

// ответ с JWT
type TokenResponse struct {
	Token string `json:"token"`
}

// ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// ответ с балансами
type BalanceResponse struct {
	Balance map[string]float64 `json:"balance"`
}

// ответ на пополнение
type DepositResponse struct {
	Message    string             `json:"message"`
	NewBalance map[string]float64 `json:"new_balance"`
}

// ответ курсы
type RatesResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// ответ на обмен
type ExchangeResponse struct {
	Message         string             `json:"message"`
	ExchangedAmount float64            `json:"exchanged_amount"`
	NewBalance      map[string]float64 `json:"new_balance"`
}
