package janice_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/stevecallear/janice"
)

func TestNewLogger(t *testing.T) {
	d := map[string]interface{}{
		"value": "expected",
	}

	b := new(bytes.Buffer)
	l := janice.NewLogger(log.New(b, "", 0), "value={{value}}")

	l.Log(d)

	if b.String() != "value=expected\n" {
		t.Errorf("Log(); got %s, expected value=expected\n", b.String())
	}
}
