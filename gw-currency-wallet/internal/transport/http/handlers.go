package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/auth"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/cache"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/kafka"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/storages"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/transport/exchanger"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	storage   storages.Storage
	auth      *auth.Service
	exchanger *exchanger.Client
	cache     *cache.RatesCache
	producer  *kafka.Producer
	logger    *slog.Logger
}

func NewHandler(storage storages.Storage, authSvc *auth.Service, exchanger *exchanger.Client, ratesCache *cache.RatesCache, producer *kafka.Producer, logger *slog.Logger) *Handler {
	return &Handler{
		storage:   storage,
		auth:      authSvc,
		exchanger: exchanger,
		cache:     ratesCache,
		producer:  producer,
		logger:    logger,
	}
}

// Register godoc
// @Summary      Регистрация пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Данные"
// @Success      201 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Router       /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request" + err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("error hashing password")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	user := &storages.User{
		ID:           uuid.NewString(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := h.storage.CreateUser(c.Request.Context(), user); err != nil {
		if errors.Is(err, storages.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Username or email already exists"})
			return
		}
		h.logger.Error("error creating user", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, MessageResponse{Message: "User registered successfully"})

}

// Login godoc
// @Summary      Авторизация
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Данные"
// @Success      200 {object} TokenResponse
// @Failure      401 {object} ErrorResponse
// @Router       /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request" + err.Error()})
		return
	}
	user, err := h.storage.GetUserByUsername(c.Request.Context(), req.Username)
	if errors.Is(err, storages.ErrNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Invalid username or password"})
		return
	}
	if err != nil {
		h.logger.Error("error getting user", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
		return
	}

	token, err := h.auth.Generate(user.ID, user.Username)
	if err != nil {
		h.logger.Error("error generating token", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// GetBalance godoc
// @Summary      Получение баланса
// @Tags         wallet
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} BalanceResponse
// @Failure      401 {object} ErrorResponse
// @Router       /balance [get]
func (h *Handler) GetBalance(c *gin.Context) {
	userId := c.GetString(userIDKey)

	balances, err := h.storage.GetBalances(c.Request.Context(), userId)
	if err != nil {
		h.logger.Error("error getting balances", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, BalanceResponse{Balance: balances})
}

// Deposit godoc
// @Summary      Пополнение счёта
// @Tags         wallet
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body DepositRequest true "Данные"
// @Success      200 {object} DepositResponse
// @Failure      400 {object} ErrorResponse
// @Router       /wallet/deposit [post]
func (h *Handler) Deposit(c *gin.Context) {
	userId := c.GetString(userIDKey)

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request" + err.Error()})
		return
	}
	balances, err := h.storage.Deposit(c.Request.Context(), userId, req.Currency, req.Amount)
	if err != nil {
		h.logger.Error("error getting balances", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, DepositResponse{Message: "Deposit successfully", NewBalance: balances})
}

// Withdraw godoc
// @Summary      Вывод средств
// @Tags         wallet
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body WithdrawRequest true "Данные"
// @Success      200 {object} DepositResponse
// @Failure      400 {object} ErrorResponse
// @Router       /wallet/withdraw [post]
func (h *Handler) Withdraw(c *gin.Context) {
	userId := c.GetString(userIDKey)

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request" + err.Error()})
		return
	}

	balances, err := h.storage.Withdraw(c.Request.Context(), userId, req.Currency, req.Amount)
	if errors.Is(err, storages.ErrInsufficientBalance) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "insufficient balance to withdraw"})
		return
	}
	if err != nil {
		h.logger.Error("failed to withdraw", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, DepositResponse{
		Message:    "Withdrawal successful",
		NewBalance: balances,
	})
}

// GetRates godoc
// @Summary      Курсы валют
// @Tags         exchange
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]map[string]float64
// @Failure      500 {object} ErrorResponse
// @Router       /exchange/rates [get]
func (h *Handler) GetRates(c *gin.Context) {
	ctx := c.Request.Context()

	rates, err := h.exchanger.GetRates(ctx)
	if err != nil {
		h.logger.Error("error getting rates", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if err := h.cache.SetRates(ctx, rates); err != nil {
		h.logger.Warn("error setting rates", "error", err)
	}

	c.JSON(http.StatusOK, RatesResponse{Rates: rates})
}

// Exchange godoc
// @Summary      Обмен валют
// @Tags         exchange
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ExchangeRequest true "Данные"
// @Success      200 {object} ExchangeResponse
// @Failure      400 {object} ErrorResponse
// @Router       /exchange [post]
func (h *Handler) Exchange(c *gin.Context) {
	userId := c.GetString(userIDKey)
	ctx := c.Request.Context()

	var req ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request" + err.Error()})
		return
	}

	if req.FromCurrency == req.ToCurrency {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency and tocurrency cannot be the same"})
		return
	}

	rates, err := h.cache.GetRates(ctx)
	if errors.Is(err, cache.ErrCacheMiss) {
		rates, err = h.exchanger.GetRates(ctx)
		if err != nil {
			h.logger.Error("error getting rates", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve exchange rates"})
			return
		}
		_ = h.cache.SetRates(ctx, rates)
	} else if err != nil {
		h.logger.Warn("error getting rates from cache", "error", err)
		rates, err = h.exchanger.GetRates(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve exchange rates"})
			return
		}
	}

	fromRate, ok := rates[req.FromCurrency]
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid currency"})
		return
	}

	toRate, ok := rates[req.ToCurrency]
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid currency"})
		return
	}

	rate := fromRate / toRate

	balances, resultAmount, err := h.storage.Exchange(ctx, userId, req.FromCurrency, req.ToCurrency, req.Amount, rate)
	if errors.Is(err, storages.ErrInsufficientBalance) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Insufficient funds or invalid currencies"})
		return
	}
	if err != nil {
		h.logger.Error("failed to exchange rates", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	if req.Amount >= 30000 {
		event := kafka.Event{
			TransactionID: uuid.NewString(),
			UserID:        userId,
			FromCurrency:  req.FromCurrency,
			ToCurrency:    req.ToCurrency,
			Amount:        req.Amount,
			Timestamp:     time.Now(),
			Status:        kafka.StatusSuccess,
		}
		if err := h.producer.Produce(c.Request.Context(), event); err != nil {
			h.logger.Error("failed to produce event", "error", err)
		}
	}

	c.JSON(http.StatusOK, ExchangeResponse{
		Message:         "Successfully exchanged rates",
		ExchangedAmount: resultAmount,
		NewBalance:      balances,
	})
}
