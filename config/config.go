package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"
)

type Database struct {
	database string
	username string
	password string
	host     string
	port     string
}

// DefaultDatabase creates and returns a Database struct with default values,
// reading from environment variables if available.
func DefaultDatabase() *Database {
	return &Database{
		database: getEnv("DB_DATABASE", "mydatabase"),
		username: getEnv("DB_USERNAME", "user"),
		password: getEnv("DB_PASSWORD", "password"),
		host:     getEnv("DB_HOST", "localhost"),
		port:     getEnv("DB_PORT", "5432"),
	}
}

func (d *Database) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		d.username, d.password, d.host, d.port, d.database,
	)
}

type APIKEY struct {
	GoogleSecret   string
	GoogleID       string
	GoogleCallBack string
}

func DefaultAPIKEY() *APIKEY {
	return &APIKEY{
		GoogleSecret:   getEnv("GOOGLE_SECRET", ""),
		GoogleID:       getEnv("GOOGLE_ID", ""),
		GoogleCallBack: getEnv("GOOGLE_CALLBACK", ""),
	}
}

type Server struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Rate limiting
	RateLimit      float64
	RateBurst      int
	RateCleanupInt time.Duration

	// Shutdown
	ShutdownTimeout time.Duration
}

func DefaultConfig() *Server {
	return &Server{
		Port:            getEnv("SERVER_PORT", "8080"),
		ReadTimeout:     parseDuration(getEnv("SERVER_READ_TIMEOUT", "10s")),
		WriteTimeout:    parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "30s")),
		IdleTimeout:     parseDuration(getEnv("SERVER_IDLE_TIMEOUT", "1m")),
		RateLimit:       parseFloat(getEnv("SERVER_RATE_LIMIT", "0.5")),
		RateBurst:       parseInt(getEnv("SERVER_RATE_BURST", "5")),
		RateCleanupInt:  parseDuration(getEnv("SERVER_RATE_CLEANUP_INT", "1m")),
		ShutdownTimeout: parseDuration(getEnv("SERVER_SHUTDOWN_TIMEOUT", "30s")),
	}
}

type AppConfig struct {
	Log          *slog.Logger
	Environement string
	HTTP         *Server
	Database     *Database
	APIKEY       *APIKEY
}

var (
	appConfigInstance *AppConfig
	once              sync.Once
)

// Load initializes and returns the AppConfig. It uses sync.Once to ensure
// the configuration is loaded only once, even if called multiple times.
func Load() *AppConfig {
	once.Do(func() {
		appConfigInstance = &AppConfig{
			Environement: getEnv("APP_ENV", "development"),
			HTTP:         DefaultConfig(),
			Database:     DefaultDatabase(),
			APIKEY:       DefaultAPIKEY(),
		}
		appConfigInstance.slogInit()
	})
	return appConfigInstance
}

func (c *AppConfig) slogInit() {
	var handler slog.Handler
	var level slog.Level

	// Determine log level from environment
	switch getEnv("LOG_LEVEL", "") {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		// Default based on environment
		if c.Environement == "production" {
			level = slog.LevelInfo
		} else {
			level = slog.LevelDebug
		}
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: c.Environement != "production", // Add source info in non-prod
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000Z07:00"))
			}
			return a
		},
	}

	logFormat := getEnv("LOG_FORMAT", "")
	if c.Environement == "production" || logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Create logger with app-wide attributes
	logger := slog.New(handler)
	slog.SetDefault(logger)
	c.Log = logger
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		fmt.Printf("Warning: Could not parse duration string '%s'. Using default.\n", s)
		return 0
	}
	return d
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Printf("Warning: Could not parse float string '%s'. Using default.\n", s)
		return 0.0
	}
	return f
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Warning: Could not parse int string '%s'. Using default.\n", s)
		return 0
	}
	return i
}
