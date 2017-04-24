package janice

import (
	"log"
	"os"

	"github.com/valyala/fasttemplate"
)

const (
	// RequestLogFormat represents a default request log format
	RequestLogFormat = `{"time":"{{time}}","log_type":"{{log_type}}","host":"{{host}}","method":"{{method}}","path":"{{path}}","code":"{{code}}","duration":"{{duration}}","written":"{{written}}"}`

	// ErrorLogFormat represents a default error log format
	ErrorLogFormat = `{"time":"{{time}}","log_type":"{{log_type}}","error":"{{error}}"}`
)

var (
	// RequestLogger represents a default request logger
	RequestLogger = NewLogger(log.New(os.Stdout, "", 0), RequestLogFormat)

	// ErrorLogger represents a default error logger
	ErrorLogger = NewLogger(log.New(os.Stderr, "", 0), ErrorLogFormat)
)

type (
	// Logger represents a formatting logger
	Logger interface {
		Log(map[string]interface{})
	}

	templateLogger struct {
		logger   *log.Logger
		template *fasttemplate.Template
	}
)

// NewLogger returns a new Logger
func NewLogger(l *log.Logger, template string) Logger {
	t := fasttemplate.New(template, "{{", "}}")

	return &templateLogger{
		logger:   l,
		template: t,
	}
}

func (l *templateLogger) Log(v map[string]interface{}) {
	s := l.template.ExecuteString(v)
	l.logger.Println(s)
}
