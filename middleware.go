package janice

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
)

// StatusError represents an error with an associated HTTP status code
type StatusError struct {
	Code int
	error
}

// NewStatusError returns a new StatusError
func NewStatusError(code int, err error) *StatusError {
	return &StatusError{
		Code:  code,
		error: err,
	}
}

// Recovery returns a recovery middleware
func Recovery(l Logger) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			defer func() {
				if r := recover(); r != nil {
					l.Log(map[string]interface{}{
						"log_type": "panic",
						"time":     time.Now().UTC().Format(time.RFC3339),
						"error":    fmt.Sprintf("%v", r),
					})

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			return h(w, r)
		}
	}
}

// RequestLogging returns a request logging middleware
func RequestLogging(l Logger) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			d := map[string]interface{}{
				"log_type": "request",
				"time":     time.Now().UTC().Format(time.RFC3339),
				"host":     r.Host,
				"method":   r.Method,
				"path":     r.URL.String(),
			}

			var err error
			wh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = h(w, r)
			})

			m := httpsnoop.CaptureMetrics(wh, w, r)

			d["code"] = strconv.Itoa(m.Code)
			d["duration"] = m.Duration.String()
			d["written"] = strconv.FormatInt(m.Written, 10)

			l.Log(d)
			return err
		}
	}
}

// ErrorHandling returns an error logging middleware
func ErrorHandling() MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			if err := h(w, r); err != nil {
				c := http.StatusInternalServerError
				if serr, ok := err.(*StatusError); ok {
					c = serr.Code
				}

				http.Error(w, err.Error(), c)
			}

			return nil
		}
	}
}

// ErrorLogging returns an error logging middleware
func ErrorLogging(l Logger) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			err := h(w, r)

			if err != nil {
				d := map[string]interface{}{
					"log_type": "error",
					"time":     time.Now().UTC().Format(time.RFC3339),
					"error":    err.Error(),
				}

				l.Log(d)
			}

			return err
		}
	}
}
