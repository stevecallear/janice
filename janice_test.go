package janice_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stevecallear/janice"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		middleware []janice.MiddlewareFunc
		handler    janice.HandlerFunc
		exp        string
	}{
		{
			name:       "should allow empty slice",
			middleware: []janice.MiddlewareFunc{},
			handler:    newHandler("a", nil),
			exp:        "a",
		},
		{
			name: "should apply middleware",
			middleware: []janice.MiddlewareFunc{
				newMiddleware("a", nil),
			},
			handler: newHandler("b", nil),
			exp:     "ab",
		},
		{
			name: "should apply middleware slice",
			middleware: []janice.MiddlewareFunc{
				newMiddleware("a", nil),
				newMiddleware("b", nil),
			},
			handler: newHandler("c", nil),
			exp:     "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			b := new(bytes.Buffer)
			h := janice.New(tt.middleware...).Then(tt.handler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(context.WithValue(req.Context(), logContextKey, b))

			h.ServeHTTP(rec, req)
			if b.String() != tt.exp {
				st.Errorf("got %s, expected %s", b.String(), tt.exp)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{
			name: "should return a nil error",
			code: http.StatusOK,
		},
		{
			name: "should wrap the handler",
			code: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
			})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			err := janice.Wrap(h)(rec, req)
			if err != nil {
				st.Errorf("got %v, expected nil", err)
			}
			if rec.Code != tt.code {
				st.Errorf("got %d, expected %d", rec.Code, tt.code)
			}
		})
	}
}

func TestMiddlewareFunc_Append(t *testing.T) {
	err := errors.New("error")
	tests := []struct {
		name       string
		middleware [][]janice.MiddlewareFunc
		handler    janice.HandlerFunc
		err        error
		exp        string
	}{
		{
			name:       "should allow nil middleware",
			middleware: [][]janice.MiddlewareFunc{{}},
			handler:    newHandler("a", nil),
			exp:        "a",
		},
		{
			name: "should append middleware",
			middleware: [][]janice.MiddlewareFunc{
				{
					newMiddleware("a", nil),
				},
				{
					newMiddleware("b", nil),
				},
			},
			handler: newHandler("c", nil),
			exp:     "abc",
		},
		{
			name: "should append middleware slice",
			middleware: [][]janice.MiddlewareFunc{
				{
					newMiddleware("a", nil),
				},
				{
					newMiddleware("b", nil),
					newMiddleware("c", nil),
				},
			},
			handler: newHandler("d", nil),
			exp:     "abcd",
		},
		{
			name: "should pass errors through",
			middleware: [][]janice.MiddlewareFunc{
				{
					newMiddleware("a", nil),
				},
				{
					newMiddleware("b", err),
					newMiddleware("c", nil),
				},
			},
			handler: newHandler("d", nil),
			err:     err,
			exp:     "abcd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			b := new(bytes.Buffer)
			var mw janice.MiddlewareFunc
			for i, fn := range tt.middleware {
				if i == 0 {
					mw = janice.New(fn...)
				} else {
					mw = mw.Append(fn...)
				}
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(context.WithValue(req.Context(), logContextKey, b))

			err := mw(tt.handler)(rec, req)
			if err != tt.err {
				st.Errorf("got %v, expected %v", err, tt.err)
			}
			if b.String() != tt.exp {
				st.Errorf("got %s, expected %s", b.String(), tt.exp)
			}
		})
	}
}

func TestMiddlewareFunc_Then(t *testing.T) {
	tests := []struct {
		name       string
		middleware janice.MiddlewareFunc
		handler    janice.HandlerFunc
		err        error
		code       int
		exp        string
	}{
		{
			name:       "should wrap the handler",
			middleware: newMiddleware("a", nil),
			handler:    newHandler("b", nil),
			code:       http.StatusOK,
			exp:        "ab",
		},
		{
			name:       "should handle errors",
			middleware: newMiddleware("a", nil),
			handler:    newHandler("b", errors.New("error")),
			code:       http.StatusInternalServerError,
			exp:        "ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			b := bytes.NewBuffer(nil)
			h := tt.middleware.Then(tt.handler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(context.WithValue(req.Context(), logContextKey, b))

			h.ServeHTTP(rec, req)
			if rec.Code != tt.code {
				st.Errorf("got %d, expected %d", rec.Code, tt.code)
			}
			if b.String() != tt.exp {
				st.Errorf("got %s, expected %s", b.String(), tt.exp)
			}
		})
	}
}

type contextKey string

var logContextKey = contextKey("log")

func newHandler(msg string, err error) janice.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) error {
		l := r.Context().Value(logContextKey).(io.Writer)
		fmt.Fprint(l, msg)
		return err
	}
}

func newMiddleware(msg string, err error) janice.MiddlewareFunc {
	return func(n janice.HandlerFunc) janice.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			l := r.Context().Value(logContextKey).(io.Writer)
			fmt.Fprint(l, msg)
			if err := n(w, r); err != nil {
				return err
			}
			return err
		}
	}
}
