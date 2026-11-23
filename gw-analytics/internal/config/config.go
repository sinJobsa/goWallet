package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Kafka      KafkaConfig
	ClickHouse ClickHouseConfig
	Server     ServerConfig
	Log        LogConfig
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type ClickHouseConfig struct {
	Addr     string
	Database string
	User     string
	Password string
	Timeout  time.Duration
}

type ServerConfig struct {
	ShutdownTimeout time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

func Load(path string) (*Config, error) {
	fileVars, err := loadEnvFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Kafka: KafkaConfig{
			Brokers: getStringSlice("KAFKA_BROKERS", fileVars, DefaultKafkaBrokers),
			Topic:   getString("KAFKA_TOPIC", fileVars, DefaultKafkaTopic),
			GroupID: getString("KAFKA_GROUP_ID", fileVars, DefaultKafkaGroupID),
		},
		ClickHouse: ClickHouseConfig{
			Addr:     getString("CLICKHOUSE_ADDR", fileVars, DefaultClickHouseAddr),
			Database: getString("CLICKHOUSE_DATABASE", fileVars, DefaultClickHouseDatabase),
			User:     getString("CLICKHOUSE_USER", fileVars, DefaultClickHouseUser),
			Password: getString("CLICKHOUSE_PASSWORD", fileVars, DefaultClickHousePassword),
			Timeout:  getDuration("CLICKHOUSE_TIMEOUT", fileVars, DefaultClickHouseTimeout),
		},
		Server: ServerConfig{
			ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", fileVars, DefaultShutdownTimeout),
		},
		Log: LogConfig{
			Level:  getString("LOG_LEVEL", fileVars, DefaultLogLevel),
			Format: getString("LOG_FORMAT", fileVars, DefaultLogFormat),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}
	return cfg, nil
}

func loadEnvFile(path string) (map[string]string, error) {
	fileVars, err := godotenv.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	return fileVars, nil
}

func getString(key string, fileVars map[string]string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if v, ok := fileVars[key]; ok && v != "" {
		return v
	}
	return def
}

func getDuration(key string, fileVars map[string]string, def time.Duration) time.Duration {
	raw := getString(key, fileVars, "")
	if raw == "" {
		return def
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return def
	}
	return v
}

func getStringSlice(key string, fileVars map[string]string, def string) []string {
	raw := getString(key, fileVars, def)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func (c *Config) Validate() error {
	if len(c.Kafka.Brokers) == 0 {
		return errors.New("KAFKA_BROKERS is required")
	}
	if c.Kafka.Topic == "" {
		return errors.New("KAFKA_TOPIC is required")
	}
	if c.Kafka.GroupID == "" {
		return errors.New("KAFKA_GROUP_ID is required")
	}
	if c.ClickHouse.Addr == "" {
		return errors.New("CLICKHOUSE_ADDR is required")
	}
	if c.ClickHouse.Database == "" {
		return errors.New("CLICKHOUSE_DATABASE is required")
	}
	if c.ClickHouse.Timeout <= 0 {
		return errors.New("CLICKHOUSE_TIMEOUT must be positive")
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be debug|info|warn|error, got %s", c.Log.Level)
	}
	switch c.Log.Format {
	case "json", "text":
	default:
		return fmt.Errorf("LOG_FORMAT must be json|text, got %s", c.Log.Format)
	}
	return nil
}
