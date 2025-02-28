// Package logger wrapper for zerolog
package logger

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Config logger settings
type Config struct {
	Level             string
	TimeFormat        string
	PrettyPrint       bool
	RedirectStdLogger bool
	DisableSampling   bool
	ErrorStack        bool
}

// Logger object capable of interacting with Logger
type Logger struct {
	zero              zerolog.Logger
	zeroErr           zerolog.Logger
	level             string
	prettyPrint       bool
	redirectSTDLogger bool
	initialized       bool
}

// New creates a new Logger
func New(config Config) *Logger {
	zerolog.SetGlobalLevel(getZeroLogLevel(config.Level))
	zerolog.DisableSampling(config.DisableSampling)
	zerolog.TimeFieldFormat = config.TimeFormat
	if config.ErrorStack {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	}

	var logger Logger
	logger.level = config.Level
	logger.prettyPrint = config.PrettyPrint
	logger.redirectSTDLogger = config.RedirectStdLogger

	logger.compile()

	return &logger
}

// NewDefault creates a new default Logger
func NewDefault() *Logger {
	return New(Config{
		Level:      "debug",
		TimeFormat: time.RFC3339,
	})
}

// Debug starts a new message with debug level
func (l *Logger) Debug() *zerolog.Event {
	return l.zero.Debug()
}

// Info starts a new message with info level
func (l *Logger) Info() *zerolog.Event {
	return l.zero.Info()
}

// Error starts a new message with error level
func (l *Logger) Error() *zerolog.Event {
	return l.zeroErr.Error()
}

// Warn starts a new message with warn level
func (l *Logger) Warn() *zerolog.Event {
	return l.zeroErr.Warn()
}

// Panic starts a new message with panic level
func (l *Logger) Panic() *zerolog.Event {
	return l.zeroErr.Panic()
}

// With creates a child logger with the field added to its context
func (l *Logger) With() zerolog.Context {
	return l.zero.With()
}

// Fatal sends the event with fatal level
func (l *Logger) Fatal(v ...interface{}) {
	l.zeroErr.Fatal().Msgf("%v", v)
}

// Fatalf sends the event with formatted msg with fatal level
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.zeroErr.Fatal().Msgf(format, v...)
}

// Print sends the event with debug level
func (l *Logger) Print(v ...interface{}) {
	l.zero.Debug().Msgf("%v", v)
}

// Printf sends the event with formatted msg with debug level
func (l *Logger) Printf(format string, v ...interface{}) {
	l.zero.Debug().Msgf(format, v...)
}

func (l *Logger) init() {
	l.initialized = true

	outWriters := []io.Writer{os.Stdout}
	errWriters := []io.Writer{os.Stderr}

	l.zero = zerolog.New(zerolog.MultiLevelWriter(outWriters...)).With().Logger()
	l.zeroErr = zerolog.New(zerolog.MultiLevelWriter(errWriters...)).With().Logger()
}

func (l *Logger) compile() {
	if !l.initialized {
		l.init()
	}

	if l.redirectSTDLogger {
		l.setLogOutputToZeroLog()
	}

	l.initDefaultFields()

	if l.prettyPrint {
		l.addPrettyPrint()
	}
}

func (l *Logger) initDefaultFields() {
	l.zero = l.zero.With().Timestamp().Logger()
	l.zeroErr = l.zeroErr.With().Timestamp().Logger()
}

func (l *Logger) addPrettyPrint() {
	prettyStdout := zerolog.ConsoleWriter{Out: os.Stdout}
	prettyStderr := zerolog.ConsoleWriter{Out: os.Stderr}

	l.zero = l.zero.Output(prettyStdout)
	l.zeroErr = l.zeroErr.Output(prettyStderr)
}

func (l *Logger) setLogOutputToZeroLog() {
	log.SetFlags(0)
	log.SetOutput(l.zero)
}

func getZeroLogLevel(lvl string) zerolog.Level {
	switch strings.ToLower(lvl) {
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
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	}
	return zerolog.NoLevel
}
