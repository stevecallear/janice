package janice

import (
	"fmt"
	"net/http"
	"strconv"

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

// StatusCode returns a status code for the specified error
// By default, nil errors return 200 while non-nil errors return 500
// If the error is non-nil and a StatusError then the code will be returned
func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if serr, ok := err.(*StatusError); ok {
		return serr.Code
	}

	return http.StatusInternalServerError
}

// Recovery returns a recovery middleware
func Recovery(l Logger) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			defer func() {
				if r := recover(); r != nil {
					l.Error(Fields{
						"type":  "recovery",
						"error": fmt.Sprint(r),
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
			// store the request url in case it is changed later in the middleware pipe
			ru := r.URL.String()

			var err error
			wh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = h(w, r)
			})

			m := httpsnoop.CaptureMetrics(wh, w, r)

			l.Info(Fields{
				"type":     "request",
				"host":     r.Host,
				"method":   r.Method,
				"path":     ru,
				"code":     strconv.Itoa(m.Code),
				"duration": m.Duration.String(),
				"written":  strconv.FormatInt(m.Written, 10),
			})

			return err
		}
	}
}

// ErrorHandling returns an error logging middleware
func ErrorHandling() MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			if err := h(w, r); err != nil {
				http.Error(w, err.Error(), StatusCode(err))
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
				l.Error(Fields{
					"type":  "error",
					"error": err.Error(),
				})
			}

			return err
		}
	}
}
