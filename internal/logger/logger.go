package logger

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var setLogOnce sync.Once
var log zerolog.Logger

var setLogLevelOnce sync.Once
var logLevel int

// SetLogLevel sets the log level to the input value.
func SetLogLevel(level int) {
	setLogLevelOnce.Do(func() {
		logLevel = level
	})
}

// Get returns a singleton instance of zerolog.Logger.
//
// It is configured to write logs to stdout in a colorized, human-friendly format.
//
// Logging is performed at the level specified by the LOG_LEVEL environment variable (see config.yaml file).
func Get() zerolog.Logger {
	setLogOnce.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339

		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Logger()
	})

	return log
}

// WithRequestId returns a logger with the request_id field set to the value of the X-Request-ID header from the input http request.
// If the header is not present, the request_id field is set to "unknown".
func WithRequestId(r *http.Request) zerolog.Logger {
	requestId, ok := r.Context().Value("X-Request-ID").(string)
	if !ok {
		requestId = "unknown"
	}

	return Get().With().Str("request_id", requestId).Logger()
}
