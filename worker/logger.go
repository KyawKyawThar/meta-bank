package worker

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logger struct {
}

func NewLogger() *logger {
	return &logger{}
}

func (logger *logger) Print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))
}

// Note: reserving value zero to differentiate unspecified case.
//level_unspecified LogLevel = iota

// Printf only for redis internal log system
func (logger *logger) Printf(ctx context.Context, format string, v ...interface{}) {
	log.WithLevel(zerolog.DebugLevel).Msgf(format, v...)
}

// Debug DebugLevel is the lowest level of logging.
// Debug logs are intended for debugging and development purposes.
func (logger *logger) Debug(args ...interface{}) {
	logger.Print(zerolog.DebugLevel, args)
}

// Info InfoLevel is used for general informational log messages.
func (logger *logger) Info(args ...interface{}) {
	logger.Print(zerolog.InfoLevel, args)
}

// Warn WarnLevel is used for undesired but relatively expected events, which may indicate a problem.
func (logger *logger) Warn(args ...interface{}) {}

// Error ErrorLevel is used for undesired and unexpected events that the program can recover from.
func (logger *logger) Error(args ...interface{}) {}

// Fatal FatalLevel is used for undesired and unexpected events that the program cannot recover from.
func (logger *logger) Fatal(args ...interface{}) {}
