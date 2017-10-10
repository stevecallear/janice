package janice_test

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stevecallear/janice"
)

func TestNewStatusError(t *testing.T) {
	tests := []struct {
		name string
		code int
		err  error
	}{
		{
			name: "should return the correct error",
			code: http.StatusInternalServerError,
			err:  errors.New("error"),
		},
		{
			name: "should return the correct code",
			code: http.StatusBadRequest,
			err:  errors.New("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			err := janice.NewStatusError(tt.code, tt.err)
			if err.Error() != tt.err.Error() {
				st.Errorf("got %s, expected %s", err.Error(), tt.err.Error())
			}
			if err.Code != tt.code {
				st.Errorf("got %d, expected %d", err.Code, tt.code)
			}
		})
	}
}

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		exp  int
	}{
		{
			name: "should return 200 if error is nil",
			exp:  http.StatusOK,
		},
		{
			name: "should return 500 if error is not nil",
			err:  errors.New("error"),
			exp:  http.StatusInternalServerError,
		},
		{
			name: "should return code if error is status error",
			err:  janice.NewStatusError(http.StatusNotFound, errors.New("error")),
			exp:  http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			act := janice.StatusCode(tt.err)
			if act != tt.exp {
				st.Errorf("got %d, expected %d", act, tt.exp)
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		panic error
		err   error
		code  int
		exp   map[string]string
	}{
		{
			name: "should do nothing if there is no panic",
			code: http.StatusOK,
			exp:  map[string]string{},
		},
		{
			name: "should do nothing if there is an error",
			err:  errors.New("error"),
			code: http.StatusOK,
			exp:  map[string]string{},
		},
		{
			name:  "should recover and log the panic",
			panic: errors.New("panic"),
			code:  http.StatusInternalServerError,
			exp: map[string]string{
				"type":  "recovery",
				"level": "error",
				"error": "panic",
				"time":  now.UTC().Format(time.RFC3339),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			withTime(now, func() {
				h := func(w http.ResponseWriter, r *http.Request) error {
					if tt.panic != nil {
						panic(tt.panic)
					}
					return tt.err
				}
				b := new(bytes.Buffer)
				l := janice.NewLogger(log.New(b, "", 0))

				rec := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/", nil)

				err := janice.Recovery(l)(h)(rec, req)
				if err != tt.err {
					st.Errorf("got %v, expected %v", err, tt.err)
				}
				if rec.Code != tt.code {
					st.Errorf("got %d, expected %d", rec.Code, tt.code)
				}
				act := parseLogEntry(b.Bytes())
				if !reflect.DeepEqual(act, tt.exp) {
					st.Errorf("got %v, expected %v", act, tt.exp)
				}
			})
		})
	}
}

func TestRequestLogging(t *testing.T) {
	type request struct {
		method string
		path   string
	}
	type result struct {
		code int
		path string
	}
	now := time.Now()
	tests := []struct {
		name string
		req  request
		res  result
		err  error
		exp  map[string]string
	}{
		{
			name: "should log the method",
			req: request{
				method: "GET",
				path:   "/",
			},
			res: result{
				code: http.StatusOK,
				path: "/",
			},
			exp: map[string]string{
				"type":    "request",
				"level":   "info",
				"time":    now.UTC().Format(time.RFC3339),
				"host":    "example.com",
				"method":  "GET",
				"path":    "/",
				"code":    strconv.Itoa(http.StatusOK),
				"written": "0",
			},
		},
		{
			name: "should log the status code",
			req: request{
				method: "POST",
				path:   "/",
			},
			res: result{
				code: http.StatusBadRequest,
				path: "/",
			},
			exp: map[string]string{
				"type":    "request",
				"level":   "info",
				"time":    now.UTC().Format(time.RFC3339),
				"host":    "example.com",
				"method":  "POST",
				"path":    "/",
				"code":    strconv.Itoa(http.StatusBadRequest),
				"written": "0",
			},
		},
		{
			name: "should log the path",
			req: request{
				method: "GET",
				path:   "/resource/",
			},
			res: result{
				code: http.StatusOK,
				path: "/",
			},
			exp: map[string]string{
				"type":    "request",
				"level":   "info",
				"time":    now.UTC().Format(time.RFC3339),
				"host":    "example.com",
				"method":  "GET",
				"path":    "/resource/",
				"code":    strconv.Itoa(http.StatusOK),
				"written": "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			withTime(now, func() {
				h := func(w http.ResponseWriter, r *http.Request) error {
					r.URL.Path = tt.res.path
					w.WriteHeader(tt.res.code)
					return nil
				}
				b := new(bytes.Buffer)
				l := janice.NewLogger(log.New(b, "", 0))

				rec := httptest.NewRecorder()
				req := httptest.NewRequest(tt.req.method, tt.req.path, nil)

				err := janice.RequestLogging(l)(h)(rec, req)
				if err != tt.err {
					st.Errorf("got %v, expected %v", err, tt.err)
				}
				if rec.Code != tt.res.code {
					st.Errorf("got %d, expected %d", rec.Code, tt.res.code)
				}
				act := parseLogEntry(b.Bytes())
				delete(act, "duration") // ignore the duration field to avoid flakiness
				if !reflect.DeepEqual(act, tt.exp) {
					st.Errorf("got %v, expected %v", act, tt.exp)
				}
			})
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		exp  string
	}{
		{
			name: "should return 200 if the error is nil",
			code: http.StatusOK,
		},
		{
			name: "should log and return 500 if the error is not nil",
			err:  errors.New("error"),
			code: http.StatusInternalServerError,
			exp:  "error\n",
		},
		{
			name: "should log and return code if the error is status error",
			err:  janice.NewStatusError(http.StatusBadRequest, errors.New("error")),
			code: http.StatusBadRequest,
			exp:  "error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			h := janice.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				return tt.err
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			err := janice.ErrorHandling()(h)(rec, req)
			if err != nil {
				t.Errorf("got %v, expected nil", err)
			}
			if rec.Code != tt.code {
				t.Errorf("got %d, expected %d", rec.Code, tt.code)
			}
			act := new(bytes.Buffer)
			act.ReadFrom(rec.Body)
			if act.String() != tt.exp {
				t.Errorf("got %s, expected %s", act.String(), tt.exp)
			}
		})
	}
}

func TestErrorLogging(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		err  error
		code int
		exp  map[string]string
	}{
		{
			name: "should not write a log entry if the error is nil",
			code: http.StatusOK,
			exp:  map[string]string{},
		},
		{
			name: "should write a log entry if the error is not nil",
			err:  errors.New("error"),
			code: http.StatusOK,
			exp: map[string]string{
				"type":  "error",
				"level": "error",
				"error": "error",
				"time":  now.UTC().Format(time.RFC3339),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			h := func(w http.ResponseWriter, r *http.Request) error {
				return tt.err
			}
			b := new(bytes.Buffer)
			l := janice.NewLogger(log.New(b, "", 0))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			err := janice.ErrorLogging(l)(h)(rec, req)
			if err != tt.err {
				st.Errorf("got %v, expected %v", err, tt.err)
			}
			if rec.Code != tt.code {
				st.Errorf("got %d, expected %d", rec.Code, tt.code)
			}
			act := parseLogEntry(b.Bytes())
			if !reflect.DeepEqual(act, tt.exp) {
				st.Errorf("got %v, expected %v", act, tt.exp)
			}
		})
	}
}
