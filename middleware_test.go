package janice_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestStatusCode(t *testing.T) {
	tests := []struct {
		err error
		exp int
	}{
		{
			exp: http.StatusOK,
		},
		{
			err: errors.New("error"),
			exp: http.StatusInternalServerError,
		},
		{
			err: janice.NewStatusError(http.StatusNotFound, errors.New("error")),
			exp: http.StatusNotFound,
		},
	}

	for tn, tt := range tests {
		act := janice.StatusCode(tt.err)

		if act != tt.exp {
			t.Errorf("StatusCode(%d); got %d, expected %d", tn, act, tt.exp)
		}
	}
}

func TestRecovery(t *testing.T) {
	tests := []struct {
		panic error
		err   error
		code  int
		log   map[string]string
	}{
		{
			code: http.StatusOK,
			log:  map[string]string{},
		},
		{
			panic: errors.New("panic"),
			code:  http.StatusInternalServerError,
			log: map[string]string{
				"type":  "recovery",
				"level": "error",
				"error": "panic",
			},
		},
		{
			err:  errors.New("error"),
			code: http.StatusOK,
			log:  map[string]string{},
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
		l := janice.NewLogger(b)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		err := janice.Recovery(l)(h)(rec, req)

		if err != tt.err {
			t.Errorf("Recovery(%d); got %v, expected %v", tn, err, tt.err)
		}
		if rec.Code != tt.code {
			t.Errorf("Recovery(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}

		e := readLogEntry(b.Bytes())
		if !e.hasValues(tt.log) {
			t.Errorf("Recovery(%d); got %v, expected %v", tn, e, tt.log)
		}
	}
}

func TestRequestLogging(t *testing.T) {
	tests := []struct {
		method string
		code   int
		rpath  string
		npath  string
		err    error
		log    map[string]string
	}{
		{
			method: "GET",
			code:   http.StatusOK,
			rpath:  "/",
			npath:  "/",
			log: map[string]string{
				"type":   "request",
				"level":  "info",
				"method": "GET",
				"path":   "/",
				"code":   strconv.Itoa(http.StatusOK),
			},
		},
		{
			method: "POST",
			code:   http.StatusBadRequest,
			rpath:  "/",
			npath:  "/",
			log: map[string]string{
				"type":   "request",
				"level":  "info",
				"method": "POST",
				"path":   "/",
				"code":   strconv.Itoa(http.StatusBadRequest),
			},
		},
		{
			method: "GET",
			code:   http.StatusOK,
			rpath:  "/resource/",
			npath:  "/",
			log: map[string]string{
				"type":   "request",
				"level":  "info",
				"method": "GET",
				"path":   "/resource/",
				"code":   strconv.Itoa(http.StatusOK),
			},
		},
	}

	for tn, tt := range tests {
		h := func(w http.ResponseWriter, r *http.Request) error {
			r.URL.Path = tt.npath

			w.WriteHeader(tt.code)
			return nil
		}

		b := new(bytes.Buffer)
		l := janice.NewLogger(b)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(tt.method, tt.rpath, nil)

		err := janice.RequestLogging(l)(h)(rec, req)

		if err != tt.err {
			t.Errorf("RequestLogging(%d); got %v, expected %v", tn, err, tt.err)
		}
		if rec.Code != tt.code {
			t.Errorf("RequestLogging(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}

		e := readLogEntry(b.Bytes())
		if !e.hasValues(tt.log) {
			t.Errorf("RequestLogging(%d); got %v, expected %v", tn, e, tt.log)
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

		err := janice.ErrorHandling()(h)(rec, req)

		b := new(bytes.Buffer)
		b.ReadFrom(rec.Body)

		if err != nil {
			t.Errorf("ErrorHandling(%d); got %v, expected nil", tn, err)
		}
		if rec.Code != tt.code {
			t.Errorf("ErrorHandling(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}
		if b.String() != tt.exp {
			t.Errorf("ErrorHandling(%d); got %s, expected %s", tn, b.String(), tt.exp)
		}
	}
}

func TestErrorLogging(t *testing.T) {
	tests := []struct {
		err  error
		code int
		log  map[string]string
	}{
		{
			code: http.StatusOK,
			log:  map[string]string{},
		},
		{
			err:  errors.New("error"),
			code: http.StatusOK,
			log: map[string]string{
				"type":  "error",
				"level": "error",
				"error": "error",
			},
		},
	}

	for tn, tt := range tests {
		h := func(w http.ResponseWriter, r *http.Request) error {
			return tt.err
		}

		b := new(bytes.Buffer)
		l := janice.NewLogger(b)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		err := janice.ErrorLogging(l)(h)(rec, req)

		if err != tt.err {
			t.Errorf("ErrorLogging(%d); got %v, expected %v", tn, err, tt.err)
		}
		if rec.Code != tt.code {
			t.Errorf("ErrorLogging(%d); got %d, expected %d", tn, rec.Code, tt.code)
		}

		e := readLogEntry(b.Bytes())
		if !e.hasValues(tt.log) {
			t.Errorf("ErrorLogging(%d); got %v, expected %v", tn, e, tt.log)
		}
	}
}
