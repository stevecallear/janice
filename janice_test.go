package janice_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stevecallear/janice"
)

func TestNew(t *testing.T) {
	err := errors.New("error")
	tests := []struct {
		name       string
		middleware []janice.MiddlewareFunc
		handlerFn  janice.HandlerFunc
		body       string
		err        error
	}{
		{
			name:       "should handle empty arguments",
			middleware: []janice.MiddlewareFunc{},
			handlerFn:  newHandler("h").handle,
			body:       "|h|",
		},
		{
			name: "should handle a single middleware func",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mw"),
			},
			handlerFn: newHandler("h").handle,
			body:      "|mw:before||h||mw:after|",
		},
		{
			name: "should handle multiple middleware funcs",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").handle,
			body:      "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
		},
		{
			name: "should propagate errors",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").withErr(err).handle,
			body:      "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
			err:       err,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := janice.New(tt.middleware...)
			rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
			if err := chain(tt.handlerFn)(rec, req); err != tt.err {
				t.Errorf("got %v, expected %v", err, tt.err)
			}
			if rec.body.String() != tt.body {
				t.Errorf("got %s, expected %s", rec.body.String(), tt.body)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	t.Run("should invoke the handler and return a nil error", func(t *testing.T) {
		const body = "body"
		h := janice.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
		h(rec, req)
		if rec.code != http.StatusOK {
			t.Errorf("got %d, expected %d", rec.code, http.StatusOK)
		}
		if rec.body.String() != body {
			t.Errorf("got %s, expected %s", rec.body.String(), body)
		}
	})
}

func TestWrapFunc(t *testing.T) {
	t.Run("should invoke the handler and return a nil error", func(t *testing.T) {
		const body = "body"
		h := janice.WrapFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		})
		rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
		h(rec, req)
		if rec.code != http.StatusOK {
			t.Errorf("got %d, expected %d", rec.code, http.StatusOK)
		}
		if rec.body.String() != body {
			t.Errorf("got %s, expected %s", rec.body.String(), body)
		}
	})
}

func TestMiddlewareFunc_Append(t *testing.T) {
	err := errors.New("error")
	tests := []struct {
		name       string
		middleware []janice.MiddlewareFunc
		handlerFn  janice.HandlerFunc
		body       string
		err        error
	}{
		{
			name: "should handle empty arguments",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mw"),
			},
			handlerFn: newHandler("h").handle,
			body:      "|mw:before||h||mw:after|",
		},
		{
			name: "should append a single middleware func",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").handle,
			body:      "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
		},
		{
			name: "should multiple middleware funcs",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
				newMiddlewareFunc("mwc"),
			},
			handlerFn: newHandler("h").handle,
			body:      "|mwa:before||mwb:before||mwc:before||h||mwc:after||mwb:after||mwa:after|",
		},
		{
			name: "should propagate errors",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").withErr(err).handle,
			body:      "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
			err:       err,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := tt.middleware[0].Append(tt.middleware[1:]...)
			rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
			if err := chain(tt.handlerFn)(rec, req); err != tt.err {
				t.Errorf("got %v, expected %v", err, tt.err)
			}
			if rec.body.String() != tt.body {
				t.Errorf("got %s, expected %s", rec.body.String(), tt.body)
			}
		})
	}
}

func TestMiddlewareFunc_Then(t *testing.T) {
	err := errors.New("error")
	tests := []struct {
		name       string
		middleware []janice.MiddlewareFunc
		handlerFn  janice.HandlerFunc
		errorFn    func(*testing.T, error)
		code       int
		body       string
	}{
		{
			name: "should apply the middleware chain",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").handle,
			body:      "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
			code:      http.StatusOK,
		},
		{
			name: "should handle errors by default",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").withErr(err).handle,
			body:      "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
			code:      http.StatusInternalServerError,
		},
		{
			name: "should pass errors to the error func",
			middleware: []janice.MiddlewareFunc{
				newMiddlewareFunc("mwa"),
				newMiddlewareFunc("mwb"),
			},
			handlerFn: newHandler("h").withErr(err).handle,
			errorFn: func(t *testing.T, e error) {
				if e != err {
					t.Errorf("got %v, expected %v", e, err)
				}
			},
			body: "|mwa:before||mwb:before||h||mwb:after||mwa:after|",
			code: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := janice.New(tt.middleware...).Then(tt.handlerFn)
			if tt.errorFn != nil {
				h.ErrorFn = func(_ http.ResponseWriter, _ *http.Request, err error) {
					tt.errorFn(t, err)
				}
			}
			rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
			h.ServeHTTP(rec, req)
			if rec.code != tt.code {
				t.Errorf("got %d, expected %d", rec.code, tt.code)
			}
			if rec.body.String() != tt.body {
				t.Errorf("got %s, expected %s", rec.body.String(), tt.body)
			}
		})
	}
}

func BenchmarkNew10(b *testing.B) {
	benchmarkNew(b, 10)
}

func BenchmarkNew100(b *testing.B) {
	benchmarkNew(b, 100)
}

func benchmarkNew(b *testing.B, n int) {
	mw := make([]janice.MiddlewareFunc, n, n)
	for i := 0; i < n; i++ {
		mw[i] = newMiddlewareFunc(fmt.Sprintf("mw%d", i))
	}
	h := janice.New(mw...).Then(newHandler("h").handle)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkAppendSingle10(b *testing.B) {
	benchmarkAppendSingle(b, 10)
}

func BenchmarkAppendSingle100(b *testing.B) {
	benchmarkAppendSingle(b, 100)
}

func benchmarkAppendSingle(b *testing.B, n int) {
	chain := janice.New()
	for i := 0; i < n; i++ {
		chain = chain.Append(newMiddlewareFunc(fmt.Sprintf("mw%d", i)))
	}
	h := chain.Then(newHandler("h").handle)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkAppendMultiple10(b *testing.B) {
	benchmarkAppendMultiple(b, 10)
}

func BenchmarkAppendMultiple100(b *testing.B) {
	benchmarkAppendMultiple(b, 100)
}

func benchmarkAppendMultiple(b *testing.B, n int) {
	mw := make([]janice.MiddlewareFunc, n, n)
	for i := 0; i < n; i++ {
		mw[i] = newMiddlewareFunc(fmt.Sprintf("mw%d", i))
	}
	h := janice.New().Append(mw...).Then(newHandler("h").handle)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec, req := newResponseRecorder(), httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(rec, req)
	}
}

type (
	responseRecorder struct {
		code int
		body *bytes.Buffer
	}
	handler struct {
		name string
		err  error
	}
)

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{
		code: http.StatusOK,
		body: bytes.NewBuffer(nil),
	}
}

func (r *responseRecorder) Header() http.Header {
	panic("not implemented")
}

func (r *responseRecorder) WriteHeader(c int) {
	r.code = c
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	return r.body.Write(p)
}

func newHandler(name string) handler {
	return handler{name: name}
}

func (h handler) withErr(err error) handler {
	h.err = err
	return h
}

func (h handler) handle(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "|%s|", h.name)
	return h.err
}

func newMiddlewareFunc(name string) janice.MiddlewareFunc {
	return func(n janice.HandlerFunc) janice.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			fmt.Fprintf(w, "|%s:before|", name)
			err := n(w, r)
			fmt.Fprintf(w, "|%s:after|", name)
			return err
		}
	}
}
