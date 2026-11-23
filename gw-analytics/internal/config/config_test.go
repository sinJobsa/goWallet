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
		Kafka: KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "large_transfers",
			GroupID: "gw-analytics",
		},
		ClickHouse: ClickHouseConfig{
			Addr:     "localhost:9000",
			Database: "analytics",
			User:     "default",
			Password: "",
			Timeout:  5 * time.Second,
		},
		Server: ServerConfig{ShutdownTimeout: 10 * time.Second},
		Log:    LogConfig{Level: "info", Format: "json"},
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeTempEnv(t, `
KAFKA_BROKERS=kafka1:9092,kafka2:9092
KAFKA_TOPIC=large_transfers
KAFKA_GROUP_ID=gw-analytics
CLICKHOUSE_ADDR=clickhouse:9000
CLICKHOUSE_DATABASE=analytics
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_TIMEOUT=5s
LOG_LEVEL=debug
LOG_FORMAT=text
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Kafka.Brokers) != 2 {
		t.Errorf("Kafka.Brokers = %v, want 2", cfg.Kafka.Brokers)
	}
	if cfg.Kafka.Brokers[0] != "kafka1:9092" {
		t.Errorf("Kafka.Brokers[0] = %q", cfg.Kafka.Brokers[0])
	}
	if cfg.Kafka.GroupID != "gw-analytics" {
		t.Errorf("Kafka.GroupID = %q", cfg.Kafka.GroupID)
	}
	if cfg.ClickHouse.Addr != "clickhouse:9000" {
		t.Errorf("ClickHouse.Addr = %q", cfg.ClickHouse.Addr)
	}
	if cfg.ClickHouse.Timeout != 5*time.Second {
		t.Errorf("ClickHouse.Timeout = %v", cfg.ClickHouse.Timeout)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.env")
	t.Setenv("CLICKHOUSE_DATABASE", "env-db")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Kafka.Topic != DefaultKafkaTopic {
		t.Errorf("Kafka.Topic = %q, want %q", cfg.Kafka.Topic, DefaultKafkaTopic)
	}
	if cfg.ClickHouse.Database != "env-db" {
		t.Errorf("ClickHouse.Database = %q, want env-db", cfg.ClickHouse.Database)
	}
}

func TestValidate_EmptyBrokers(t *testing.T) {
	cfg := validConfig()
	cfg.Kafka.Brokers = nil
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for empty KAFKA_BROKERS")
	}
}

func TestValidate_EmptyGroupID(t *testing.T) {
	cfg := validConfig()
	cfg.Kafka.GroupID = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for empty KAFKA_GROUP_ID")
	}
}

func TestValidate_EmptyClickHouseAddr(t *testing.T) {
	cfg := validConfig()
	cfg.ClickHouse.Addr = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for empty CLICKHOUSE_ADDR")
	}
}

func TestValidate_EmptyClickHouseDatabase(t *testing.T) {
	cfg := validConfig()
	cfg.ClickHouse.Database = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for empty CLICKHOUSE_DATABASE")
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := validConfig()
	cfg.Log.Level = "verbose"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate: expected error for invalid LOG_LEVEL")
	}
}
