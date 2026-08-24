package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

type LoggerOptions struct {
	Level      string
	LogFile    string
	JSONFormat bool
}

func InitLogger(opts ...LoggerOptions) {
	var output io.Writer = os.Stderr
	var err error

	option := LoggerOptions{Level: "info"}
	if len(opts) > 0 {
		option = opts[0]
	}

	lvl, err := zerolog.ParseLevel(strings.ToLower(option.Level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	if option.LogFile != "" {
		output, err = os.OpenFile(option.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v, using stderr\n", err)
			output = os.Stderr
		}
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        output,
		TimeFormat: time.RFC3339,
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%s", i))
		},
	}

	if option.JSONFormat || option.LogFile != "" {
		Log = zerolog.New(output).
			Level(lvl).
			With().
			Timestamp().
			CallerWithSkipFrameCount(3).
			Logger()
	} else {
		Log = zerolog.New(consoleWriter).
			Level(lvl).
			With().
			Timestamp().
			CallerWithSkipFrameCount(3).
			Logger()
	}
}

func InitDefault() {
	InitLogger(LoggerOptions{Level: "info"})
}

func WithLevel(level string) zerolog.Logger {
	lvl, _ := zerolog.ParseLevel(strings.ToLower(level))
	return Log.Level(lvl)
}

func With() zerolog.Context {
	return Log.With()
}

func Debug() *zerolog.Event {
	return Log.Debug()
}

func Info() *zerolog.Event {
	return Log.Info()
}

func Warn() *zerolog.Event {
	return Log.Warn()
}

func Error() *zerolog.Event {
	return Log.Error()
}

func Fatal() *zerolog.Event {
	return Log.Fatal()
}

func Panic() *zerolog.Event {
	return Log.Panic()
}
