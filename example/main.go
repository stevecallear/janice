package main

import (
	"fmt"
	"net/http"

	"github.com/stevecallear/janice"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", janice.New(middleware).Then(func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "hello handler!\n")
		return nil
	}))

	http.ListenAndServe(":8080", janice.Default().Then(janice.Wrap(mux)))
}

func middleware(n janice.HandlerFunc) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "hello middleware!\n")
		return n(w, r)
	}
}
