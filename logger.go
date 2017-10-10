package janice

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	// Now returns the current time
	Now = func() time.Time {
		return time.Now()
	}

	// DefaultLogger is the default logger
	DefaultLogger = NewLogger(log.New(os.Stdout, "", 0))
)

type (
	// Fields represents a set of log fields
	Fields map[string]interface{}

	// Logger represents a basic structured logger
	Logger interface {
		Info(Fields)
		Error(Fields)
	}
	logger struct {
		*log.Logger
	}
)

// NewLogger returns a new logger
func NewLogger(l *log.Logger) Logger {
	return &logger{
		Logger: l,
	}
}

// Info logs the specified fields at info level
func (l *logger) Info(f Fields) {
	l.log("info", f)
}

// Error logs the specified fields at error level
func (l *logger) Error(f Fields) {
	l.log("error", f)
}

func (l *logger) log(level string, f Fields) {
	f["level"] = level
	f["time"] = Now().UTC().Format(time.RFC3339)
	b, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}
	l.Println(string(b))
}
