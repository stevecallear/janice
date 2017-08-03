package janice_test

import (
	"bytes"
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
			t.Errorf("New(%d); got %s, expected %s", tn, b.String(), tt.exp)
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
		mwm [][]string
		msg string
		err error
		exp string
	}{
		{
			mwm: [][]string{{}},
			msg: "a",
			exp: "a",
		},
		{
			mwm: [][]string{{"a"}, {"b"}},
			msg: "c",
			exp: "abc",
		},
		{
			mwm: [][]string{{"a"}, {"b", "c"}},
			msg: "d",
			exp: "abcd",
		},
		{
			mwm: [][]string{{"a"}, {"b", "c"}},
			msg: "d",
			err: errors.New("error"),
			exp: "abcd",
		},
	}

	for tn, tt := range tests {
		b := new(bytes.Buffer)

		var mw janice.MiddlewareFunc
		for i, ms := range tt.mwm {
			if i == 0 {
				mw = janice.New(newMiddlewareSlice(b, ms)...)
			} else {
				mw = mw.Append(newMiddlewareSlice(b, ms)...)
			}
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		err := mw(newHandler(b, tt.msg, tt.err))(rec, req)

		if err != tt.err {
			t.Errorf("Append(%d); got %v, expected %v", tn, err, tt.err)
		}
		if b.String() != tt.exp {
			t.Errorf("Append(%d); got %s, expected %s", tn, b.String(), tt.exp)
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
			mwm:  "a",
			msg:  "b",
			code: http.StatusOK,
			exp:  "ab",
		},
		{
			mwm:  "a",
			msg:  "b",
			err:  errors.New("error"),
			code: http.StatusInternalServerError,
			exp:  "ab",
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

func newMiddlewareSlice(log io.Writer, msgs []string) []janice.MiddlewareFunc {
	fns := []janice.MiddlewareFunc{}
	for _, msg := range msgs {
		fns = append(fns, newMiddleware(log, msg))
	}
	return fns
}

func newMiddleware(log io.Writer, msg string) janice.MiddlewareFunc {
	return func(n janice.HandlerFunc) janice.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			fmt.Fprint(log, msg)
			return n(w, r)
		}
	}
}
