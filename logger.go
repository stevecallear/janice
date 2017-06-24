package janice

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// DefaultLogger is the default logger
var DefaultLogger = NewLogger(log.New(os.Stdout, "", 0))

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

func (l *logger) Info(f Fields) {
	l.log("info", f)
}

func (l *logger) Error(f Fields) {
	l.log("error", f)
}

func (l *logger) log(level string, f Fields) {
	f["level"] = level
	f["time"] = time.Now().UTC().Format(time.RFC3339)

	b, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}

	l.Println(string(b))
}
