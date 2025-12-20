package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func Init(level string, pretty bool) {
	// set logging level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	if pretty {
		log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func Get() *zerolog.Logger {
	return &log
}

// return a looger with component
func WithComponent(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}
