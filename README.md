# Janice
[![Build Status](https://github.com/stevecallear/janice/actions/workflows/build.yml/badge.svg)](https://github.com/stevecallear/janice/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/stevecallear/janice/branch/master/graph/badge.svg)](https://codecov.io/gh/stevecallear/janice)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevecallear/janice)](https://goreportcard.com/report/github.com/stevecallear/janice)

Janice simplifies middleware chaining in Go HTTP applications. It is heavily inspired by both [Alice](https://github.com/justinas/alice) and [Negroni](https://github.com/urfave/negroni), but with a focus on simplifying handler functions by removing error handling boilerplate.

## Getting started
```
go get github.com/stevecallear/janice
```
```
http.ListenAndServe(":8080", janice.New(middlewareFunc).Then(handlerFunc))
```

## Handlers
Janice uses the following signature for HTTP handler functions. This allows error handling to be moved outside of handler functions and into common middleware:
```
func(w http.ResponseWriter, r *http.Request) error
```

## Middleware
Middleware functions simply return a handler that wraps in input function with additional logic, for example:
```
func middleware(n janice.HandlerFunc) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Println("before handler")
		err := n(w, r)
		log.Println("after handler")
		return err
	}
}
```

Middleware can be combined using either the `New` or `Append` functions, for example:
```
a := janice.New(m1, m2)
b := a.Append(m3)
c := a.Append(m4, m5)
```

Middleware chains can be converted to an `http.Handler` implementation using the `Then` function. By default any errors that make it through the middleware chain will result in `http.StatusInternalServerError` being written to the response. It is possible to customise this using the `ErrorFn` property. For example, the following will log any errors, leaving the status code unchanged:
```
h := janice.New(middlewareFunc).Then(handlerFunc)
h.ErrorFn = func(_ http.ResponseWriter, _ *http.Request, err error) {
	log.Printf("error: %v", err)
}
```

`janice.Wrap` and `janice.WrapFunc` allow `http.Handler` and `http.HandlerFunc` implementations to be used in middleware chains. For example:
```
mux := http.NewServeMux()
mux.Handle("/", handler)
janice.New(middleware).Then(janice.Wrap(mux))
```

## Example
The following example creates two middleware chains: the first is applied to the route and handles any errors returned by the handler function, while the second is applied to the mux to log incoming requests.
```
package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/stevecallear/janice"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", janice.New(errorHandling).Then(func(w http.ResponseWriter, r *http.Request) error {
		if e := r.URL.Query().Get("err"); e != "" {
			return errors.New(e)
		}
		return nil
	}))
	http.ListenAndServe(":8080", janice.New(requestLogging).Then(janice.Wrap(mux)))
}

func requestLogging(n janice.HandlerFunc) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Printf("%s %s\n", r.Method, r.URL.Path)
		return n(w, r)
	}
}

func errorHandling(n janice.HandlerFunc) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if err := n(w, r); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return nil
	}
}
```