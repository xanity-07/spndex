package loggerpkg

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/zerologWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	pkgerrors "github.com/rs/zerolog/pkgerrors"
	"github.com/xanity-07/spndex/internal/config"
)

// LoggerService manages logger creation and New Relic integration
type LoggerService struct {
	nrApp *newrelic.Application
}

// NewLoggerService created a new LoggerService with New Relic integration
// with all our observavility configurations
func NewLoggerService(cfg *config.ObservabilityConfig) *LoggerService {
	service := &LoggerService{}

	// Check if license key is provided if not return a empty service no observability
	if cfg.NewRelic.LicenseKey == "" {
		fmt.Println("New Relic license key not provided, skipping initialization")
		return service
	}

	// Configure New Relic to forward logs and tracers to New Relic
	var configOptions []newrelic.ConfigOption
	configOptions = append(configOptions,
		newrelic.ConfigAppName(cfg.ServiceName),
		newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(cfg.NewRelic.AppLogsForwardingEnabled),
		newrelic.ConfigDistributedTracerEnabled(cfg.NewRelic.DistributedTracingEnabled),
	)

	// Add DebugLogger only if explicitly enabled
	// Disabled for both local and production environment enable if there are issues with New Relic
	if cfg.NewRelic.DebugLogger {
		configOptions = append(configOptions,
			newrelic.ConfigDebugLogger(os.Stdout),
		)
	}

	// Initialize New Relic a application and spread configOptions
	app, err := newrelic.NewApplication(configOptions...)
	if err != nil {
		fmt.Printf("Failed to initialize New Relic: %s\n", err)
		return service
	}

	service.nrApp = app
	fmt.Printf("New Relic application initialized for app: %s\n", cfg.ServiceName)
	return service
}

// Shutdown terminated the instance of out New Relic application
func (ls *LoggerService) Shutdown() {
	if ls.nrApp != nil {
		ls.nrApp.Shutdown(time.Second * 10)
	}
}

// GetApplication returns the New Relic application instance
func (ls *LoggerService) GetApplication() *newrelic.Application {
	return ls.nrApp
}

// NewLogger creates a new logger with specified level (backwards compatible)
func NewLogger(level string, isProd bool) zerolog.Logger {
	return NewLoggerWithService(&config.ObservabilityConfig{
		Logging: config.LoggingConfig{
			Level: level,
		},
		Environment: func() string {
			if isProd {
				return "production"
			}
			return "development"
		}(),
	}, nil)
}

// NewLoggerWithService created a new logger with full config and logger service
func NewLoggerWithService(cfg *config.ObservabilityConfig, loggerService *LoggerService) zerolog.Logger {
	var logLevel zerolog.Level
	level := cfg.GetLogLevel()

	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	// Don't set global level - let each logger have its own level
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer

	// This block of code deals with whether we should forward our logs to New Relic or not
	// setup base writer
	var baseWriter io.Writer
	if cfg.IsProduction() && cfg.Logging.Format == "json" {
		// production structured logs to stdout
		baseWriter = os.Stdout

		// Wrap with New Relic zerologWriter for log forwarding in production
		if loggerService != nil && loggerService.GetApplication() != nil {
			nrWriter := zerologWriter.New(baseWriter, loggerService.nrApp)
			writer = nrWriter
		} else {
			writer = baseWriter
		}
	} else {
		// Development mode - use console writer
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	// Pass the writer and log level
	logger := zerolog.New(writer).
		Level(logLevel).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()

	// Include stack traces for errors in development
	if !cfg.IsProduction() {
		logger.With().Stack().Logger()
	}

	return logger
}

// NewLoggerWithConfig created a new logger with full config (backwards compatible)
func NewLoggerWithConfig(cfg *config.ObservabilityConfig) zerolog.Logger {
	return NewLoggerWithService(cfg, nil)
}

// WithTraceContext adds New Relic transaction context to logger
func WithTraceContext(logger zerolog.Logger, txn *newrelic.Transaction) zerolog.Logger {
	if txn == nil {
		return logger
	}

	// Get trace metadata transaction
	metadata := txn.GetTraceMetadata()

	return logger.With().
		Str("trace.id", metadata.TraceID).
		Str("span.id", metadata.SpanID).
		Logger()
}

// NewPgxLogger creates a database logger
func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatFieldValue: func(i interface{}) string {
			switch v := i.(type) {
			case string:
				// Clean and format SQL for better readability
				if len(v) > 200 {
					// Truncate very long SQL statements
					return v[:200] + "..."
				}
				return v

			case []byte:
				// Format the args for the SQL statements
				var obj interface{}
				if err := json.Unmarshal(v, obj); err != nil {
					pretty, _ := json.MarshalIndent(obj, "", "    ")
					return "\n" + string(pretty)
				}
				return string(v)

			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}

	return zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("component", "database").
		Logger()
}

// GetPgxTraceLogLevel converts zerolog level to pgx tracelog level
func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.DebugLevel:
		return 6 // tracelog.LogLevelDebug
	case zerolog.InfoLevel:
		return 4 // tracelog.LogLevelInfo
	case zerolog.WarnLevel:
		return 3 // tracelog.LogLevelWarn
	case zerolog.ErrorLevel:
		return 2 // tracelog.LogLevelError
	default:
		return 0 // tracelog.LogLevelNone
	}
}
