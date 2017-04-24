package janice

import "net/http"

type (
	// HandlerFunc represents an HTTP handler func
	HandlerFunc func(http.ResponseWriter, *http.Request) error

	// MiddlewareFunc represents a middleware func
	MiddlewareFunc func(HandlerFunc) HandlerFunc
)

// Default returns a default middleware pipe
func Default() MiddlewareFunc {
	return New(
		Recovery(ErrorLogger),
		RequestLogging(RequestLogger),
		ErrorHandling(),
		ErrorLogging(ErrorLogger))
}

// New returns a new middleware pipe
func New(m ...MiddlewareFunc) MiddlewareFunc {
	if len(m) < 1 {
		return func(h HandlerFunc) HandlerFunc {
			return h
		}
	}

	r := m[0]
	for _, v := range m[1:] {
		r = r.Append(v)
	}

	return r
}

// Wrap wraps the specified handler, allowing it to be used with middleware
func Wrap(h http.Handler) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}

// Append appends the specified middleware funcs to the pipe
func (m MiddlewareFunc) Append(n MiddlewareFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return m(n(h))
	}
}

// Then terminates the middleware pipe with the specified handler func
func (m MiddlewareFunc) Then(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := m(h)(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
