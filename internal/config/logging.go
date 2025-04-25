package config

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"log/slog"
	"os"
	"time"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo           = "info"
	LogLevelWarn           = "warn"
	LogLevelError          = "error"
)

var (
	logLevelToSlog = map[LogLevel]slog.Level{
		LogLevelDebug: slog.LevelDebug,
		LogLevelInfo:  slog.LevelInfo,
		LogLevelWarn:  slog.LevelWarn,
		LogLevelError: slog.LevelError,
	}
	logLevelToZerolog = map[LogLevel]zerolog.Level{
		LogLevelDebug: zerolog.DebugLevel,
		LogLevelInfo:  zerolog.InfoLevel,
		LogLevelWarn:  zerolog.WarnLevel,
		LogLevelError: zerolog.ErrorLevel,
	}
)

func SetLogLevel(logLevel LogLevel) {
	zerolog.SetGlobalLevel(logLevelToZerolog[logLevel])
	slogLevel := logLevelToSlog[logLevel]
	logger := slog.New(slogzerolog.Option{Level: slogLevel, Logger: &log.Logger}.NewZerologHandler())
	slog.SetDefault(logger)
}

func SetupLogger(logLevel LogLevel) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly})
	SetLogLevel(logLevel)
}
