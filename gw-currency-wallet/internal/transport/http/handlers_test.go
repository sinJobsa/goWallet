package http_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/auth"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/storages"
	transporthttp "github.com/sinJobsa/goWallet/gw-currency-wallet/internal/transport/http"
)

type mockStorage struct {
	users    map[string]*storages.User
	balances map[string]map[string]float64
	// ...
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		users:    make(map[string]*storages.User),
		balances: make(map[string]map[string]float64),
	}
}

func (m *mockStorage) CreateUser(ctx context.Context, user *storages.User) error {
	if _, ok := m.users[user.Username]; ok {
		return storages.ErrAlreadyExists
	}
	m.users[user.Username] = user
	m.balances[user.ID] = map[string]float64{"USD": 0, "RUB": 0, "EUR": 0}
	return nil
}

func (m *mockStorage) GetUserByUsername(ctx context.Context, username string) (*storages.User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, storages.ErrNotFound
	}
	return u, nil
}

func (m *mockStorage) GetUserByID(ctx context.Context, id string) (*storages.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, storages.ErrNotFound
}

func (m *mockStorage) GetBalances(ctx context.Context, userID string) (map[string]float64, error) {
	balances, ok := m.balances[userID]
	if !ok {
		return nil, storages.ErrNotFound
	}
	// копия, чтобы тесты не мутировали внутреннее состояние
	result := make(map[string]float64, len(balances))
	for k, v := range balances {
		result[k] = v
	}
	return result, nil
}

func (m *mockStorage) GetBalance(ctx context.Context, userID, currency string) (float64, error) {
	balances, ok := m.balances[userID]
	if !ok {
		return 0, storages.ErrNotFound
	}
	amount, ok := balances[currency]
	if !ok {
		return 0, storages.ErrNotFound
	}
	return amount, nil
}

func (m *mockStorage) Deposit(ctx context.Context, userID, currency string, amount float64) (map[string]float64, error) {
	balances, ok := m.balances[userID]
	if !ok {
		return nil, storages.ErrNotFound
	}
	balances[currency] += amount
	return m.GetBalances(ctx, userID)
}

func (m *mockStorage) Withdraw(ctx context.Context, userID, currency string, amount float64) (map[string]float64, error) {
	balances, ok := m.balances[userID]
	if !ok {
		return nil, storages.ErrNotFound
	}
	if balances[currency] < amount {
		return nil, storages.ErrInsufficientBalance
	}
	balances[currency] -= amount
	return m.GetBalances(ctx, userID)
}

func (m *mockStorage) Exchange(ctx context.Context, userID, from, to string, amount, rate float64) (map[string]float64, float64, error) {
	balances, ok := m.balances[userID]
	if !ok {
		return nil, 0, storages.ErrNotFound
	}
	if balances[from] < amount {
		return nil, 0, storages.ErrInsufficientBalance
	}
	resultAmount := amount * rate
	balances[from] -= amount
	balances[to] += resultAmount
	newBalances, _ := m.GetBalances(ctx, userID)
	return newBalances, resultAmount, nil
}

func (m *mockStorage) GetTransaction(ctx context.Context, id string) (*storages.Transaction, error) {
	// если хранишь transactions — ищи
	// если не хранишь — верни ErrNotFound
	return nil, storages.ErrNotFound
}

func (m *mockStorage) Close() {
	// ничего не делаем — мок в памяти
}

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := newMockStorage()
	authSvc := auth.New("test-secret-at-least-32-characters-long", time.Hour)
	handler := transporthttp.NewHandler(storage, authSvc, nil, nil, nil, slog.Default())

	router := gin.New()
	router.POST("/api/v1/register", handler.Register)

	body := `{"username":"alice","password":"password123","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestRegister_Duplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := newMockStorage()
	authSvc := auth.New("test-secret-at-least-32-characters-long", time.Hour)
	handler := transporthttp.NewHandler(storage, authSvc, nil, nil, nil, slog.Default())

	router := gin.New()
	router.POST("/api/v1/register", handler.Register)

	body := `{"username":"alice","password":"password123","email":"alice@example.com"}`

	// --- Первая регистрация: 201 ---
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first register: status = %d, want 201, body = %s", w1.Code, w1.Body.String())
	}

	// --- Вторая регистрация (тот же username): 400 ---
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("duplicate register: status = %d, want 400, body = %s", w2.Code, w2.Body.String())
	}

	// --- Проверяем сообщение ---
	var resp transporthttp.ErrorResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error != "Username or email already exists" {
		t.Errorf("error = %q, want %q", resp.Error, "Username or email already exists")
	}
}

func TestGetBalance_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := auth.New("test-secret-at-least-32-characters-long", time.Hour)
	handler := transporthttp.NewHandler(newMockStorage(), authSvc, nil, nil, nil, slog.Default())

	router := gin.New()
	protected := router.Group("")
	protected.Use(transporthttp.JWTMiddleware(authSvc))
	protected.GET("/api/v1/balance", handler.GetBalance)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/balance", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetBalance_WithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage := newMockStorage()
	authSvc := auth.New("test-secret-at-least-32-characters-long", time.Hour)
	handler := transporthttp.NewHandler(storage, authSvc, nil, nil, nil, slog.Default())

	// создаём пользователя
	user := &storages.User{ID: "user-123", Username: "alice"}
	_ = storage.CreateUser(context.Background(), user)

	// генерируем токен
	token, _ := authSvc.Generate(user.ID, user.Username)

	router := gin.New()
	protected := router.Group("")
	protected.Use(transporthttp.JWTMiddleware(authSvc))
	protected.GET("/api/v1/balance", handler.GetBalance)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/balance", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}
