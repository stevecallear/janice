package janice_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	"github.com/stevecallear/janice"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		fn    func(janice.Logger)
		keys  []string
		vals  map[string]string
		panic bool
	}{
		{
			fn: func(l janice.Logger) {
				l.Info(janice.Fields{
					"expected": "value",
				})
			},
			keys: []string{"level", "time", "expected"},
			vals: map[string]string{
				"level":    "info",
				"expected": "value",
			},
		},
		{
			fn: func(l janice.Logger) {
				l.Error(janice.Fields{
					"expected": "value",
				})
			},
			keys: []string{"level", "time", "expected"},
			vals: map[string]string{
				"level":    "error",
				"expected": "value",
			},
		},
		{
			fn: func(l janice.Logger) {
				l.Info(janice.Fields{
					"expected": math.Inf, // force panic
				})
			},
			keys:  []string{},
			vals:  map[string]string{},
			panic: true,
		},
	}

	for tn, tt := range tests {
		b := new(bytes.Buffer)

		func() {
			defer func() {
				r := recover()
				if tt.panic && r == nil {
					t.Errorf("Logger(%d); got nil, expected an error", tn)
				}
				if !tt.panic && r != nil {
					t.Errorf("Logger(%d); got %v, expected nil", tn, r)
				}
			}()

			tt.fn(janice.NewLogger(b))
		}()

		e := readLogEntry(b.Bytes())
		if !e.hasKeys(tt.keys) {
			t.Errorf("Logger(%d); got %v, expected %v", tn, e, tt.keys)
		}
		if !e.hasValues(tt.vals) {
			t.Errorf("Logger(%d); got %v, expected %v", tn, e, tt.vals)
		}
	}
}

type logEntry map[string]string

func readLogEntry(b []byte) logEntry {
	e := logEntry{}
	json.Unmarshal(b, &e)
	return e
}

func (e logEntry) hasKeys(ks []string) bool {
	for _, k := range ks {
		if _, ok := e[k]; !ok {
			return false
		}
	}
	return len(e) == len(ks)
}

func (e logEntry) hasValues(vm map[string]string) bool {
	for k, v := range vm {
		sv, ok := e[k]
		if !ok || sv != v {
			return false
		}
	}
	return true
}
