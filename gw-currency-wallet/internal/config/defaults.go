package config

import "time"

const (
	//HTTP
	DefaultHTTPHost = "0.0.0.0"
	DefaultHTTPPort = 8080

	//Postgres
	DefaultDBHOST           = "localhost"
	DefaultDBPort           = 5432
	DefaultDBUser           = "postgres"
	DefaultDBPassword       = ""
	DefaultDBName           = "wallet"
	DefaultSSLMode          = "disable"
	DefaultDBMaxConnections = 10
	DefaultDBMinConnections = 2
	DefaultDBConnectTimeout = time.Second * 5

	//gRPC-клиент к Exchanger'у
	DefaultExchangerAddress = "localhost:50051"
	DefaultExchangerTimeout = time.Second * 2

	//Greceful shutdown
	DefaultShutdownTime = time.Second * 10

	//JWT
	DefaultJWTSecret = ""
	DefaultJWTTTL    = 24 * time.Hour

	//Redis
	DefaultRedisHost     = "localhost"
	DefaultRedisPort     = 6379
	DefaultRedisPassword = ""
	DefaultRedisDB       = 0
	DefaultRedisTTL      = 60 * time.Second

	//Kafka
	DefaultKafkaBrokers = "localhost:9092"
	DefaultKafkaTopic   = "large_transfers"

	//logs
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"
)
