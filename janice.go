package janice

import (
	"net/http"
)

type (
	// Handler represents an HTTP handler
	Handler struct {
		ErrorFn   func(http.ResponseWriter, *http.Request, error)
		handlerFn HandlerFunc
	}

	// HandlerFunc represents an HTTP handler func
	HandlerFunc func(http.ResponseWriter, *http.Request) error

	// MiddlewareFunc represents an HTTP middleware func
	MiddlewareFunc func(HandlerFunc) HandlerFunc
)

// New returns a new middleware chain containing the specified functions
func New(m ...MiddlewareFunc) MiddlewareFunc {
	switch len(m) {
	case 0:
		return func(h HandlerFunc) HandlerFunc {
			return h
		}
	case 1:
		return m[0]
	default:
		return merge(m)
	}
}

// Wrap wraps the specified HTTP handler
func Wrap(h http.Handler) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}

// WrapFunc wraps the specified HTTP handler function
func WrapFunc(h http.HandlerFunc) HandlerFunc {
	return Wrap(h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.handlerFn(w, r); err != nil {
		h.ErrorFn(w, r, err)
	}
}

// Append appends the specified middleware function to the chain
func (m MiddlewareFunc) Append(n ...MiddlewareFunc) MiddlewareFunc {
	if len(n) < 1 {
		return m
	}
	return merge(append([]MiddlewareFunc{m}, n...))
}

// Then wraps the specified handler with the middleware chain
func (m MiddlewareFunc) Then(h HandlerFunc) *Handler {
	return &Handler{
		ErrorFn: func(w http.ResponseWriter, _ *http.Request, err error) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		handlerFn: m(h),
	}
}

func merge(m []MiddlewareFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		for i := len(m) - 1; i >= 0; i-- {
			h = m[i](h)
		}
		return h
	}
}
