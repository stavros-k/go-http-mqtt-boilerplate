package config

import (
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/dialect"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type EnvKey string

const (
	EnvGenerate EnvKey = "GENERATE"

	EnvPort      EnvKey = "PORT"
	EnvDataDir   EnvKey = "DATA_DIR"
	EnvLogLevel  EnvKey = "LOG_LEVEL"
	EnvLogToFile EnvKey = "LOG_TO_FILE"

	EnvDBHost    EnvKey = "DB_HOST"
	EnvDBPort    EnvKey = "DB_PORT"
	EnvDBName    EnvKey = "DB_NAME"
	EnvDBUser    EnvKey = "DB_USER"
	EnvDBPass    EnvKey = "DB_PASSWORD"
	EnvDBSSLMode EnvKey = "DB_SSLMODE"

	EnvMQTTBrokerPort EnvKey = "MQTT_SERVER_PORT"

	EnvMQTTBroker   EnvKey = "MQTT_BROKER"
	EnvMQTTClientID EnvKey = "MQTT_CLIENT_ID"
	EnvMQTTUsername EnvKey = "MQTT_USERNAME"
	EnvMQTTPassword EnvKey = "MQTT_PASSWORD"
)

type Config struct {
	Port      int
	Generate  bool
	DataDir   string
	Database  string
	Dialect   dialect.Dialect
	LogLevel  slog.Leveler
	LogOutput io.Writer

	// MQTT Server configuration
	MQTTBrokerPort int

	// MQTT configuration
	MQTTBroker   string
	MQTTClientID string
	MQTTUsername string
	MQTTPassword string
}

func New(dbDialect dialect.Dialect) (*Config, error) {
	// Get data directory
	dataDir := getStringEnv(EnvDataDir, "data")

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Derive paths from data directory
	logPath := filepath.Join(dataDir, "app.log")

	var logOutput io.Writer = os.Stdout

	if getBoolEnv(EnvLogToFile, false) {
		f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		logOutput = f
	}

	if err := dbDialect.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database dialect: %w", err)
	}

	// Build database connection string based on dialect
	var dbConnString string

	switch dbDialect {
	case dialect.SQLite:
		// TODO: Set a common set of PRAGMA settings for SQLite connections
		dbConnString = filepath.Join(dataDir, "database.sqlite")
	case dialect.PostgreSQL:
		host := getStringEnv(EnvDBHost, "localhost")
		port := getIntEnv(EnvDBPort, 5432)
		dbName := getStringEnv(EnvDBName, "cloud")
		user := getStringEnv(EnvDBUser, "cloud")
		password := getStringEnv(EnvDBPass, "")
		sslmode := getStringEnv(EnvDBSSLMode, "disable")

		dbConnString = fmt.Sprintf(
			"postgresql://%s:%s@%s/%s?sslmode=%s",
			url.QueryEscape(user),
			url.QueryEscape(password),
			net.JoinHostPort(host, strconv.Itoa(port)),
			dbName, sslmode,
		)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbDialect)
	}

	return &Config{
		Port:           getIntEnv(EnvPort, 8080),
		Generate:       getBoolEnv(EnvGenerate, false),
		DataDir:        dataDir,
		Database:       dbConnString,
		Dialect:        dbDialect,
		LogLevel:       getLogLevelEnv(EnvLogLevel, slog.LevelInfo),
		LogOutput:      logOutput,
		MQTTBrokerPort: getIntEnv(EnvMQTTBrokerPort, 1883),
		MQTTBroker:     getStringEnv(EnvMQTTBroker, "tcp://127.0.0.1:1883"),
		MQTTClientID:   getStringEnv(EnvMQTTClientID, "http-mqtt-boilerplate-server"),
		MQTTUsername:   getStringEnv(EnvMQTTUsername, ""),
		MQTTPassword:   getStringEnv(EnvMQTTPassword, ""),
	}, nil
}

func (c *Config) Close() error {
	if f, ok := c.LogOutput.(*os.File); ok {
		if f != os.Stdout && f != os.Stderr {
			return f.Close()
		}
	}

	return nil
}

func getStringEnv(key EnvKey, defaultVal string) string {
	val, exists := os.LookupEnv(string(key))
	if !exists {
		return defaultVal
	}

	return val
}

func getBoolEnv(key EnvKey, defaultVal bool) bool {
	val, exists := os.LookupEnv(string(key))
	if !exists {
		return defaultVal
	}

	val = strings.ToLower(val)
	switch val {
	case "true", "1":
		return true
	default:
		return false
	}
}

func getIntEnv(key EnvKey, defaultVal int) int {
	val, exists := os.LookupEnv(string(key))
	if !exists {
		return defaultVal
	}

	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}

	return defaultVal
}

func getLogLevelEnv(key EnvKey, defaultVal slog.Leveler) slog.Leveler {
	val, exists := os.LookupEnv(string(key))
	if !exists {
		return defaultVal
	}

	switch strings.ToUpper(val) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	}

	return defaultVal
}
