package config

import "time"

const (
	DefaultKafkaBrokers = "localhost:9092"
	DefaultKafkaTopic   = "large_transfers"
	DefaultKafkaGroupID = "gw-analytics" // своя группа, не gw-notification

	DefaultClickHouseAddr     = "localhost:9000" // нативный порт, не HTTP 8123
	DefaultClickHouseDatabase = "analytics"
	DefaultClickHouseUser     = "default"
	DefaultClickHousePassword = ""
	DefaultClickHouseTimeout  = 5 * time.Second

	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"

	DefaultShutdownTimeout = 10 * time.Second
)
