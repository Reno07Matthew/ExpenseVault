package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	DBType        string
	SQLitePath    string
	MySQLDSN      string
	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
	JWTSecret     string
	ServerPort    int
}

// LoadConfig loads configuration from .env file and environment variables.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DBType:     getEnv("DB_TYPE", "sqlite"),
		SQLitePath: getEnv("SQLITE_PATH", getDBPath()),
		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
	}

	config.MySQLHost = getEnv("MYSQL_HOST", "localhost")
	config.MySQLPort = getEnvInt("MYSQL_PORT", 3306)
	config.MySQLUser = getEnv("MYSQL_USER", "root")
	config.MySQLPassword = getEnv("MYSQL_PASSWORD", "")
	config.MySQLDatabase = getEnv("MYSQL_DATABASE", "expensevault")
	config.ServerPort = getEnvInt("SERVER_PORT", 8080)

	mysqlDSN := getEnv("MYSQL_DSN", "")
	if mysqlDSN == "" {
		config.MySQLDSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.MySQLUser,
			config.MySQLPassword,
			config.MySQLHost,
			config.MySQLPort,
			config.MySQLDatabase,
		)
	} else {
		config.MySQLDSN = mysqlDSN
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getDBPath returns the default SQLite database path (internal use).
func getDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "expense.db"
	}
	dbDir := filepath.Join(homeDir, ".expensevault")
	_ = os.MkdirAll(dbDir, 0755)
	return filepath.Join(dbDir, "expense.db")
}
