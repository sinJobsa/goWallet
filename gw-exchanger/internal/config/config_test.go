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
			Host: "localhost", Port: 5432, User: "postgres",
			Password: "pass", Name: "exchanger", SSLMode: "disable",
			MaxConns: 10, MinConns: 2, ConnectTimeout: 5 * time.Second,
		},
		GRPC:   GRPCConfig{Host: "0.0.0.0", Port: 50051},
		Server: ServerConfig{ShutdownTimeout: 10 * time.Second},
		Log:    LogConfig{Level: "info", Format: "json"},
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeTempEnv(t, `
DB_HOST=db.example.com
DB_PORT=5432
DB_PASSWORD=secret
DB_NAME=exchanger
GRPC_HOST=0.0.0.0
GRPC_PORT=50051
LOG_LEVEL=debug
LOG_FORMAT=text
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DB.Host != "db.example.com" {
		t.Errorf("DB.Host = %q", cfg.DB.Host)
	}
	if cfg.DB.Password != "secret" {
		t.Errorf("DB.Password = %q", cfg.DB.Password)
	}
	if cfg.GRPC.Port != 50051 {
		t.Errorf("GRPC.Port = %d", cfg.GRPC.Port)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.env")
	t.Setenv("DB_PASSWORD", "env-pass")
	t.Setenv("GRPC_PORT", "50052")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DB.Host != DefaultDBHOST {
		t.Errorf("DB.Host = %q, want %q", cfg.DB.Host, DefaultDBHOST)
	}
	if cfg.DB.Password != "env-pass" {
		t.Errorf("DB.Password = %q, want env-pass", cfg.DB.Password)
	}
	if cfg.GRPC.Port != 50052 {
		t.Errorf("GRPC.Port = %d, want 50052", cfg.GRPC.Port)
	}
}

func TestValidate_MissingPassword(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Password = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for empty DB_PASSWORD")
	}
}

func TestValidate_InvalidDBPort(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for invalid DB_PORT")
	}
}

func TestValidate_InvalidGRPCPort(t *testing.T) {
	cfg := validConfig()
	cfg.GRPC.Port = 70000
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for invalid GRPC_PORT")
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := validConfig()
	cfg.Log.Level = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for invalid LOG_LEVEL")
	}
}
