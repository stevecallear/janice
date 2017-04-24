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
	return composite(m)
}

// Wrap wraps the specified handler, allowing it to be used with middleware
func Wrap(h http.Handler) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}

// Append appends the specified middleware funcs to the pipe
func (m MiddlewareFunc) Append(n ...MiddlewareFunc) MiddlewareFunc {
	return composite(append([]MiddlewareFunc{m}, n...))
}

// Then terminates the middleware pipe with the specified handler func
func (m MiddlewareFunc) Then(h HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := m(h)(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func composite(fns []MiddlewareFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		for i := len(fns) - 1; i >= 0; i-- {
			h = fns[i](h)
		}

		return h
	}
}
