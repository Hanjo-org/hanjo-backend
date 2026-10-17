package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config defines logger configuration options.
type Config struct {
	Env        string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

var Log *zap.Logger

// Init initializes the Zap logger with optional file rotation and console output.
func Init(cfg Config) *zap.Logger {
	level := zapcore.InfoLevel
	if cfg.Env == "development" {
		level = zapcore.DebugLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var cores []zapcore.Core

	// Always support console output in development or when no log file path is specified
	consoleEncoderCfg := zap.NewDevelopmentEncoderConfig()
	consoleEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)
	cores = append(cores, consoleCore)

	if cfg.FilePath != "" {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		})
		fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), fileWriter, level)
		cores = append(cores, fileCore)
	}

	Log = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
	return Log
}

// Sync flushes buffered log entries.
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
