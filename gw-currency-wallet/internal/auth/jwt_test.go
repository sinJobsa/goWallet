package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParse(t *testing.T) {
	svc := New("test-secret-at-least-32-characters-long", time.Hour)

	token, err := svc.Generate("user-123", "alice")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Username != "alice" {
		t.Errorf("Username = %q, want %q", claims.Username, "alice")
	}
}

func TestParse_ExpiredToken(t *testing.T) {
	svc := New("test-secret-at-least-32-characters-long", -time.Hour) // TTL в прошлом

	token, err := svc.Generate("user-123", "alice")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	_, err = svc.Parse(token)
	if err != ErrExpiredToken {
		t.Errorf("Parse error = %v, want ErrExpiredToken", err)
	}
}

func TestParse_InvalidToken(t *testing.T) {
	svc := New("test-secret-at-least-32-characters-long", time.Hour)

	_, err := svc.Parse("not-a-valid-token")
	if err != ErrInvalidToken {
		t.Errorf("Parse error = %v, want ErrInvalidToken", err)
	}
}

func TestParse_WrongSecret(t *testing.T) {
	svc1 := New("secret-one-at-least-32-characters-long", time.Hour)
	svc2 := New("secret-two-at-least-32-characters-long", time.Hour)

	token, err := svc1.Generate("user-123", "alice")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	_, err = svc2.Parse(token)
	if err != ErrInvalidToken {
		t.Errorf("Parse error = %v, want ErrInvalidToken", err)
	}
}

func TestParse_AlgNoneAttack(t *testing.T) {
	svc := New("test-secret-at-least-32-characters-long", time.Hour)

	// Создаём токен с alg=none — атака
	claims := Claims{
		UserID:   "hacker",
		Username: "hacker",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}

	_, err = svc.Parse(signed)
	if err != ErrInvalidToken {
		t.Errorf("Parse error = %v, want ErrInvalidToken (alg=none attack)", err)
	}
}
