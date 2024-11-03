package goorm

import (
	"log/slog"
	"os"
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
	logLevel := &slog.LevelVar{}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	logLevel.Set(slog.LevelInfo)


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
