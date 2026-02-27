package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func InitLogger(level string) {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	Log = zerolog.New(output).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()
}

func init() {
	InitLogger("info")
}
