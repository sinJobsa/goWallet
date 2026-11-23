package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp env: %v", err)
	}
	return path
}

func validConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host: "localhost", Port: 5432, Password: "pass",
		},
		HTTP:      HTTPConfig{Host: "0.0.0.0", Port: 8080},
		Exchanger: ExchangerConfig{Addr: "localhost:50051", Timeout: time.Second},
		JWT:       JWTConfig{Secret: "test-secret-at-least-32-characters-long", TTL: time.Hour},
		Redis:     RedisConfig{Host: "localhost", Port: 6379},
		Kafka:     KafkaConfig{Brokers: []string{"localhost:9092"}, Topic: "t"},
		Log:       LogConfig{Level: "info", Format: "json"},
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeTempEnv(t, `
DB_HOST=db.example.com
DB_PORT=5432
DB_PASSWORD=secret
DB_NAME=wallet
JWT_SECRET=test-secret-at-least-32-characters-long
JWT_TTL=1h
REDIS_HOST=redis.example.com
REDIS_PORT=6379
KAFKA_BROKERS=kafka1:9092,kafka2:9092
KAFKA_TOPIC=large_transfers
EXCHANGER_ADDR=exchanger:50051
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DB.Host != "db.example.com" {
		t.Errorf("DB.Host = %q", cfg.DB.Host)
	}
	if cfg.JWT.TTL != time.Hour {
		t.Errorf("JWT.TTL = %v", cfg.JWT.TTL)
	}
	if len(cfg.Kafka.Brokers) != 2 {
		t.Errorf("Kafka.Brokers = %v", cfg.Kafka.Brokers)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.env")
	t.Setenv("DB_PASSWORD", "env-pass")
	t.Setenv("JWT_SECRET", "env-secret-at-least-32-characters-long")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DB.Host != DefaultDBHOST {
		t.Errorf("DB.Host = %q", cfg.DB.Host)
	}
	if cfg.DB.Password != "env-pass" {
		t.Errorf("DB.Password = %q", cfg.DB.Password)
	}
}

func TestValidate_MissingPassword(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Password = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestValidate_ShortJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = "short"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := validConfig()
	cfg.Log.Level = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error")
	}
}
