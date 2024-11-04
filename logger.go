package goorm

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type Logger interface {
	Info(message string, args ...any)
	Error(message string, args ...any)
	Debug(message string, args ...any)
	Warn(message string, args ...any)
}

type logger struct {
	logger *slog.Logger
}

func NewDefaultLogger() Logger {
	w := os.Stderr

	defaultLogger := slog.New(tint.NewHandler(w, nil))

	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	return &logger{
		logger: defaultLogger,
	}
}

func (l *logger) Info(message string, args ...any) {
	l.logger.Info(message, args...)
}

func (l *logger) Error(message string, args ...any) {
	l.logger.Error(message, args...)
}

func (l *logger) Warn(message string, args ...any) {
	l.logger.Warn(message, args...)
}

func (l *logger) Debug(message string, args ...any) {
	l.logger.Debug(message)
}
