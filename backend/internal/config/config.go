package config

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type envKey string

const (
	envGenerate envKey = "GENERATE"

	envPort      envKey = "PORT"
	envDataDir   envKey = "DATA_DIR"
	envLogLevel  envKey = "LOG_LEVEL"
	envLogToFile envKey = "LOG_TO_FILE"

	envDBHost    envKey = "DB_HOST"
	envDBPort    envKey = "DB_PORT"
	envDBName    envKey = "DB_NAME"
	envDBUser    envKey = "DB_USER"
	envDBPass    envKey = "DB_PASSWORD"
	envDBSSLMode envKey = "DB_SSLMODE"

	envMQTTBroker   envKey = "MQTT_BROKER"
	envMQTTClientID envKey = "MQTT_CLIENT_ID"
	envMQTTUsername envKey = "MQTT_USERNAME"
	envMQTTPassword envKey = "MQTT_PASSWORD"
)

const (
	// Log rotation settings
	logMaxSize    = 500 // megabytes per log file
	logMaxBackups = 100 // number of old log files to retain
	logMaxAge     = 7   // days to retain old log files
)

type Config struct {
	Port      int
	Generate  bool
	DataDir   string
	Database  string
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

func New() (*Config, error) {
	// Get data directory
	dataDir := getStringEnv(envDataDir, "data")

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Derive paths from data directory
	logPath := filepath.Join(dataDir, "app.log")

	var logOutput io.Writer = os.Stdout

	if getBoolEnv(envLogToFile, false) {
		logOutput = &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAge:     logMaxAge,
			Compress:   true,
		}
	}

	// Build PostgreSQL connection string

	dbConnString := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=%s",
		url.QueryEscape(getStringEnv(envDBUser, "postgres")),
		url.QueryEscape(getStringEnv(envDBPass, "postgres")),
		net.JoinHostPort(
			getStringEnv(envDBHost, "localhost"),
			strconv.Itoa(getIntEnv(envDBPort, 5432)),
		),
		url.PathEscape(getStringEnv(envDBName, "postgres")),
		url.QueryEscape(getStringEnv(envDBSSLMode, "disable")),
	)

	return &Config{
		Generate: getBoolEnv(envGenerate, false),

		Port:     getIntEnv(envPort, 8080),
		DataDir:  dataDir,
		Database: dbConnString,

		LogLevel:  getLogLevelEnv(envLogLevel, slog.LevelInfo),
		LogOutput: logOutput,

		MQTTBroker:   getStringEnv(envMQTTBroker, "tcp://127.0.0.1:1883"),
		MQTTClientID: getStringEnv(envMQTTClientID, "http-mqtt-boilerplate-server"),
		MQTTUsername: getStringEnv(envMQTTUsername, ""),
		MQTTPassword: getStringEnv(envMQTTPassword, ""),
	}, nil
}

func (c *Config) Close() error {
	if f, ok := c.LogOutput.(*os.File); ok {
		if f != os.Stdout && f != os.Stderr {
			return f.Close()
		}
	}

	if l, ok := c.LogOutput.(*lumberjack.Logger); ok {
		return l.Close()
	}

	return nil
}

func getStringEnv(key envKey, defaultVal string) string {
	val, exists := os.LookupEnv(string(key))
	if !exists {
		return defaultVal
	}

	return val
}

func getBoolEnv(key envKey, defaultVal bool) bool {
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

func getIntEnv(key envKey, defaultVal int) int {
	val, exists := os.LookupEnv(string(key))
	if !exists {
		return defaultVal
	}

	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}

	return defaultVal
}

func getLogLevelEnv(key envKey, defaultVal slog.Leveler) slog.Leveler {
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
