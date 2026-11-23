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
			GroupID: "gw-notification",
		},
		Mongo: MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "notification",
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
KAFKA_GROUP_ID=gw-notification
MONGO_URI=mongodb://mongo:27017
MONGO_DATABASE=notification
MONGO_TIMEOUT=5s
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
	if cfg.Kafka.GroupID != "gw-notification" {
		t.Errorf("Kafka.GroupID = %q", cfg.Kafka.GroupID)
	}
	if cfg.Mongo.Database != "notification" {
		t.Errorf("Mongo.Database = %q", cfg.Mongo.Database)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.env")
	t.Setenv("KAFKA_GROUP_ID", "env-group")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Kafka.Topic != DefaultKafkaTopic {
		t.Errorf("Kafka.Topic = %q, want %q", cfg.Kafka.Topic, DefaultKafkaTopic)
	}
	if cfg.Kafka.GroupID != "env-group" {
		t.Errorf("Kafka.GroupID = %q", cfg.Kafka.GroupID)
	}
}

func TestValidate_EmptyBrokers(t *testing.T) {
	cfg := validConfig()
	cfg.Kafka.Brokers = nil
	if err := cfg.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestValidate_EmptyGroupID(t *testing.T) {
	cfg := validConfig()
	cfg.Kafka.GroupID = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestValidate_EmptyMongoURI(t *testing.T) {
	cfg := validConfig()
	cfg.Mongo.URI = ""
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
