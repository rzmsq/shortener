package logger

import (
	"log/slog"
	"os"
)

type Interface interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Logger struct {
	logger *slog.Logger
}

func New(env string) *Logger {
	var level slog.Level
	switch env {
	case "development":
		level = slog.LevelDebug
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	return &Logger{slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})),
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}
