package janice_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stevecallear/janice"
)

func TestNewStatusError(t *testing.T) {
	tests := []struct {
		code int
		err  error
	}{
		{
			code: http.StatusInternalServerError,
			err:  errors.New("error"),
		},
		{
			code: http.StatusBadRequest,
			err:  errors.New("error"),
		},
	}

	for tn, tt := range tests {
		err := janice.NewStatusError(tt.code, tt.err)

		if err.Code != tt.code {
			t.Errorf("NewStatusError(%d); got %d, expected %d", tn, err.Code, tt.code)
		}
		if err.Error() != tt.err.Error() {
			t.Errorf("NewStatusError(%d); got %s, expected %s", tn, err.Error(), tt.err.Error())
		}
	}
}

func TestRecovery(t *testing.T) {
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
			code: http.StatusOK,
		},
	}

	for tn, tt := range tests {
		h := func(w http.ResponseWriter, r *http.Request) error {
			if tt.panic != nil {
				panic(tt.panic)
			}

			if tt.err != nil {
				return tt.err
			}

			return nil
		}

		b := new(bytes.Buffer)
		l := janice.NewLogger(log.New(b, "", 0), "{{error}}")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		err := janice.Recovery(l)(h)(rec, req)

		if err != tt.err {
			t.Errorf("Recovery(%d); got %v, expected %v", tn, err, tt.err)
		}
		if rec.Code != tt.code {
			t.Errorf("Recovery(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
		if tt.panic != nil && b.String() != tt.panic.Error()+"\n" {
			t.Errorf("Recovery(%d); got %s, expected %s", tn, b.String(), tt.panic.Error())
		}
	}
}

func TestRequestLogging(t *testing.T) {
	tests := []struct {
		method string
		code   int
		err    error
		exp    string
	}{
		{
			method: "GET",
			code:   http.StatusOK,
			exp:    fmt.Sprintf("GET,%d\n", http.StatusOK),
		},
		{
			method: "POST",
			code:   http.StatusBadRequest,
			exp:    fmt.Sprintf("POST,%d\n", http.StatusBadRequest),
		},
	}

	for tn, tt := range tests {
		h := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(tt.code)
			return nil
		}

		b := new(bytes.Buffer)
		l := janice.NewLogger(log.New(b, "", 0), "{{method}},{{code}}")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(tt.method, "/", nil)

		err := janice.RequestLogging(l)(h)(rec, req)

		if err != tt.err {
			t.Errorf("RequestLogging(%d); got %v, expected %v", tn, err, tt.err)
		}
		if rec.Code != tt.code {
			t.Errorf("RequestLogging(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
		if b.String() != tt.exp {
			t.Errorf("RequestLogging(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		err  error
		code int
		exp  string
	}{
		{
			code: http.StatusOK,
		},
		{
			err:  errors.New("error"),
			code: http.StatusInternalServerError,
			exp:  "error\n",
		},
		{
			err:  janice.NewStatusError(http.StatusBadRequest, errors.New("error")),
			code: http.StatusBadRequest,
			exp:  "error\n",
		},
	}

	for tn, tt := range tests {
		h := janice.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return tt.err
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		janice.ErrorHandling().Then(h).ServeHTTP(rec, req)

		if rec.Code != tt.code {
			t.Errorf("ErrorHandling(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}

		b := new(bytes.Buffer)
		b.ReadFrom(rec.Body)

		if b.String() != tt.exp {
			t.Errorf("ErrorHandling(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}

func TestErrorLogging(t *testing.T) {
	tests := []struct {
		err  error
		code int
		exp  string
	}{
		{
			code: http.StatusOK,
		},
		{
			err:  errors.New("error"),
			code: http.StatusInternalServerError,
			exp:  "error\n",
		},
	}

	for tn, tt := range tests {
		h := func(w http.ResponseWriter, r *http.Request) error {
			return tt.err
		}

		b := new(bytes.Buffer)
		l := janice.NewLogger(log.New(b, "", 0), "{{error}}")

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		janice.ErrorLogging(l).Then(h).ServeHTTP(rec, req)

		if rec.Code != tt.code {
			t.Errorf("ErrorLogging(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
		if b.String() != tt.exp {
			t.Errorf("ErrorLogging(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}
