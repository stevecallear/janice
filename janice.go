package janice

import "net/http"

type (
	// HandlerFunc represents an HTTP handler func
	HandlerFunc func(http.ResponseWriter, *http.Request) error

	// MiddlewareFunc represents a middleware func
	MiddlewareFunc func(HandlerFunc) HandlerFunc
)

// New returns a new middleware pipe
func New(m ...MiddlewareFunc) MiddlewareFunc {
	return merge(m)
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
	return merge(append([]MiddlewareFunc{m}, n...))
}

// Then terminates the middleware pipe with the specified handler func
func (m MiddlewareFunc) Then(h HandlerFunc) http.Handler {
	h = m(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func merge(m []MiddlewareFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		for i := len(m) - 1; i >= 0; i-- {
			h = m[i](h)
		}
		return h
	}
}
