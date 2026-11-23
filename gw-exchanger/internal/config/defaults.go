package config

import "time"

const (
	//Postgres
	DefaultDBHOST           = "localhost"
	DefaultDBPort           = 5432
	DefaultDBUser           = "postgres"
	DefaultDBPassword       = ""
	DefaultDBName           = "exchanger"
	DefaultSSLMode          = "disable"
	DefaultDBMaxConnections = 10
	DefaultDBMinConnections = 2
	DefaultDBConnectTimeout = time.Second * 5

	//gRPC
	DefaultGRPCPort = 50051
	DefaultGRPCHost = "0.0.0.0"

	DefaultShutdownTime = time.Second * 10

	//logs
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"
)
