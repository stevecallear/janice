package main

import (
	"fmt"
	"net/http"

	"github.com/stevecallear/janice"
)

func main() {
	// create a default handler pipe
	hp := janice.New(janice.ErrorHandling(), janice.ErrorLogging(janice.ErrorLogger))

	mux := http.NewServeMux()
	mux.Handle("/", hp.Append(middleware).Then(func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "hello handler!\n")
		return nil
	}))

	// create a default mux pipe
	mp := janice.New(janice.Recovery(janice.ErrorLogger), janice.RequestLogging(janice.RequestLogger))

	http.ListenAndServe(":8080", mp.Then(janice.Wrap(mux)))
}

func middleware(n janice.HandlerFunc) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "hello middleware!\n")
		return n(w, r)
	}
}
