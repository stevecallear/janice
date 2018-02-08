# Janice
[![Build Status](https://travis-ci.org/stevecallear/janice.svg?branch=master)](https://travis-ci.org/stevecallear/janice)
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

## Example
The following example creates two middleware chains: the first is applied to the route and handles any errors returned by the handler function, while the second is applied to the mux to log incoming requests.

Each call to `Then` converts the middleware chain to an `http.Handler` implementation allowing it to be used with the standard HTTP packages, or third party structures such as routers. Conversely `janice.Wrap` can be used to convert an `http.Handler` implementation to a `janice.HandlerFunc`, allowing it to be used with a middleware chain.

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