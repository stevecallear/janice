package janice_test

import (
	"bytes"
	"encoding/json"
	"log"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stevecallear/janice"
)

func TestLogger(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		fn    func(janice.Logger)
		panic bool
		exp   map[string]string
	}{
		{
			name: "should write info level logs",
			fn: func(l janice.Logger) {
				l.Info(janice.Fields{
					"expected": "value",
				})
			},
			exp: map[string]string{
				"level":    "info",
				"expected": "value",
				"time":     now.UTC().Format(time.RFC3339),
			},
		},
		{
			name: "should write error level logs",
			fn: func(l janice.Logger) {
				l.Error(janice.Fields{
					"expected": "value",
				})
			},
			exp: map[string]string{
				"level":    "error",
				"expected": "value",
				"time":     now.UTC().Format(time.RFC3339),
			},
		},
		{
			name: "should panic when value cannot be serialized",
			fn: func(l janice.Logger) {
				l.Info(janice.Fields{
					"expected": math.Inf, // force panic
				})
			},
			panic: true,
			exp:   map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(st *testing.T) {
			withTime(now, func() {
				defer func() {
					err := recover()
					if tt.panic && err == nil {
						st.Error("got nil, expected an error")
					}
					if !tt.panic && err != nil {
						st.Errorf("got %v, expected nil", err)
					}
				}()
				b := new(bytes.Buffer)
				tt.fn(janice.NewLogger(log.New(b, "", 0)))
				act := parseLogEntry(b.Bytes())
				if !reflect.DeepEqual(act, tt.exp) {
					st.Errorf("got %v, expected %v", act, tt.exp)
				}
			})
		})
	}
}

func withTime(t time.Time, fn func()) {
	tfn := janice.Now
	defer func() {
		janice.Now = tfn
	}()
	janice.Now = func() time.Time {
		return t
	}
	fn()
}

func parseLogEntry(b []byte) map[string]string {
	m := map[string]string{}
	if len(b) < 1 {
		return m
	}
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	return m
}
