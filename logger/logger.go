package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	logs "github.com/rs/zerolog/log"
	"strings"
	"time"
)

func init() {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

// InitializeLogger initializes the global zerolog instance.
func InitializeLogger(logLevel string) {
	parsedLogLevel := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(parsedLogLevel)
	logs.Info().Msgf("current log level: %s", parsedLogLevel.String())
}

// parseLogLevel changes log level from the config to lower case and then to Level type. Returns (zerolog.Level)
func parseLogLevel(level string) zerolog.Level {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetLogger returns a logger instance associated with the given gin context.
// If the logger does not exist, a no-op logger is returned.
func GetLogger(c *gin.Context) zerolog.Logger {
	if logger, exists := c.Get("logger"); exists {
		if l, ok := logger.(zerolog.Logger); ok {
			return l
		}
	}
	return zerolog.Nop()
}
