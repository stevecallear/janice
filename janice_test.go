package janice_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stevecallear/janice"
)

func TestDefault(t *testing.T) {
	rl := janice.RequestLogger
	el := janice.ErrorLogger

	defer func() {
		janice.RequestLogger = rl
		janice.ErrorLogger = el
	}()

	b := new(bytes.Buffer)

	janice.RequestLogger = janice.NewLogger(log.New(b, "", 0), "")
	janice.ErrorLogger = janice.NewLogger(log.New(b, "", 0), "")

	tests := []struct {
		panic error
		err   error
		code  int
	}{
		{
			code: http.StatusOK,
		},
		{
			panic: errors.New("panic"),
			code:  http.StatusInternalServerError,
		},
		{
			err:  errors.New("error"),
			code: http.StatusInternalServerError,
		},
		{
			err:  janice.NewStatusError(http.StatusNotFound, errors.New("error")),
			code: http.StatusNotFound,
		},
	}

	for tn, tt := range tests {
		h := janice.Default().Then(func(w http.ResponseWriter, r *http.Request) error {
			if tt.panic != nil {
				panic(tt.panic)
			}

			return tt.err
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		h.ServeHTTP(rec, req)

		if rec.Code != tt.code {
			t.Errorf("Default(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		mwm []string
		msg string
		exp string
	}{
		{
			mwm: []string{},
			msg: "a",
			exp: "a",
		},
		{
			mwm: []string{"a"},
			msg: "b",
			exp: "ab",
		},
		{
			mwm: []string{"a", "b"},
			msg: "c",
			exp: "abc",
		},
	}

	for tn, tt := range tests {
		b := new(bytes.Buffer)

		mw := []janice.MiddlewareFunc{}
		for _, msg := range tt.mwm {
			mw = append(mw, newMiddleware(b, msg))
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		janice.New(mw...).Then(newHandler(b, tt.msg, nil)).ServeHTTP(rec, req)

		if b.String() != tt.exp {
			t.Errorf("Handler(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		code int
	}{
		{
			code: http.StatusOK,
		},
		{
			code: http.StatusNotFound,
		},
	}

	for tn, tt := range tests {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.code)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		err := janice.Wrap(h)(rec, req)

		if err != nil {
			t.Errorf("Wrap(%d); got %v, expected nil", tn, err)
		}
		if rec.Code != tt.code {
			t.Errorf("Wrap(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
	}
}

func TestMiddlewareAppend(t *testing.T) {
	tests := []struct {
		mwm []string
		msg string
		err error
		exp string
	}{
		{
			mwm: []string{"m"},
			msg: "h",
			exp: "mh",
		},
		{
			mwm: []string{"a", "b"},
			msg: "h",
			exp: "abh",
		},
		{
			mwm: []string{"a", "b"},
			msg: "h",
			err: errors.New("error"),
			exp: "abh",
		},
	}

	for tn, tt := range tests {
		b := new(bytes.Buffer)

		var mw janice.MiddlewareFunc
		for i, msg := range tt.mwm {
			if i == 0 {
				mw = newMiddleware(b, msg)
			} else {
				mw = mw.Append(newMiddleware(b, msg))
			}
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		err := mw(newHandler(b, tt.msg, tt.err))(rec, req)

		if err != tt.err {
			t.Errorf("Append(%d); got %v, expected %v", tn, err, tt.err)
		}
		if b.String() != tt.exp {
			t.Errorf("MiddlewareFunc(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}

func TestMiddlewareThen(t *testing.T) {
	tests := []struct {
		mwm  string
		msg  string
		err  error
		code int
		exp  string
	}{
		{
			mwm:  "m",
			msg:  "h",
			code: http.StatusOK,
			exp:  "mh",
		},
		{
			mwm:  "m",
			msg:  "h",
			err:  errors.New("error"),
			code: http.StatusInternalServerError,
			exp:  "mh",
		},
	}

	for tn, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		b := new(bytes.Buffer)

		h := newMiddleware(b, tt.mwm).Then(newHandler(b, tt.msg, tt.err))
		h.ServeHTTP(rec, req)

		if rec.Code != tt.code {
			t.Errorf("Then(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
		if b.String() != tt.exp {
			t.Errorf("Then(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}

func newHandler(log io.Writer, msg string, err error) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(log, msg)
		return err
	}
}

func newMiddleware(log io.Writer, msg string) janice.MiddlewareFunc {
	return func(n janice.HandlerFunc) janice.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			fmt.Fprint(log, msg)
			return n(w, r)
		}
	}
}
