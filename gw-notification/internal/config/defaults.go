package config

import "time"

const (
	// Kafka (consumer)
	DefaultKafkaBrokers = "localhost:9092"
	DefaultKafkaTopic   = "large_transfers"
	DefaultKafkaGroupID = "gw-notification"

	// Mongo
	DefaultMongoURI      = "mongodb://localhost:27017"
	DefaultMongoDatabase = "notification"
	DefaultMongoTimeout  = 5 * time.Second

	// Логирование
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"

	// Graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second
)
