package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a new configured zerolog logger.
func New(environment string, debug bool) zerolog.Logger {
	// Use pretty console output for development
	if environment == "development" {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		level := zerolog.InfoLevel
		if debug {
			level = zerolog.DebugLevel
		}
		return zerolog.New(output).
			Level(level).
			With().
			Timestamp().
			Caller().
			Str("app", "hanjo").
			Logger()
	}

	// Use JSON output for production
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}
	return zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Str("app", "hanjo").
		Logger()
}
