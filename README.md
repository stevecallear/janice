# Janice
[![Build Status](https://travis-ci.org/stevecallear/janice.svg?branch=master)](https://travis-ci.org/stevecallear/janice)
[![codecov](https://codecov.io/gh/stevecallear/janice/branch/master/graph/badge.svg)](https://codecov.io/gh/stevecallear/janice)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevecallear/janice)](https://goreportcard.com/report/github.com/stevecallear/janice)

Janice provides a set of Go HTTP middleware functions to simplify the process of building web applications. It is heavily inspired by both [Alice](https://github.com/justinas/alice) and [Negroni](https://github.com/urfave/negroni), hence the name.

## Getting Started

### Installation
```
go get github.com/stevecallear/janice
```

### Usage
```
package main

import (
	"fmt"
	"net/http"

	"github.com/stevecallear/janice"
)

func main() {
	// create a default handler pipe
	hp := janice.New(janice.ErrorHandling(), janice.ErrorLogging(janice.DefaultLogger))

	mux := http.NewServeMux()
	mux.Handle("/", hp.Append(middleware).Then(func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintf(w, "hello handler!\n")
		return nil
	}))

	// create a default mux pipe
	mp := janice.New(janice.Recovery(janice.DefaultLogger), janice.RequestLogging(janice.DefaultLogger))

	http.ListenAndServe(":8080", mp.Then(janice.Wrap(mux)))
}

func middleware(n janice.HandlerFunc) janice.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "hello middleware!\n")
		return n(w, r)
	}
}
```

## Middleware
Janice middleware has the following signature:
```
func (next janice.HandlerFunc) janice.HandlerFunc
```
With `janice.HandlerFunc` having the signature:
```
func (w http.ResponseWriter, r *http.Request) error
```
This is commonly used in Go web apps to move error handling boilerplate outside of the handler func. The `janice.Wrap` function can be used to convert a `http.Handler` implementations into a `janice.HandlerFunc`.

### Error Logging
Error logging writes any handler errors to the default error logger. The error itself is passed upwards to the next function in the pipe. A new error logging middleware function can be created using `janice.ErrorLogging()` with an appropriate logger.

### Error Handling
Error handling handles any errors returned by the inner handlers. By default, the error text will be written to the response along with an `http.StatusInternalServerError` status code. If a `janice.StatusError` is returned, then the associated status code will be used:
```
func myHandler(_ http.ResponseWriter, _ *http.Request) error {
	return janice.NewStatusError(http.StatusNotFound, errors.New("not found"))
}
```
A new error handling middleware function can be created using `janice.ErrorHandling()`.

### Request Logging
Request logging writes the logs the current request metrics. Metrics are obtained using [httpsnoop](https://github.com/felixge/httpsnoop) and log entries are written to a JSON templated request logger by default. New request logging middleware functions can be created using `janice.RequestLogging()` with an appropriate logger.

### Recovery
Recovery recovers from any panic that occurs throughout the middleware pipe. In the event of a panic the information will written to an error logger and a `http.StatusInternalServerError` code will be written to the response. New recovery middleware functions can be created using `janice.Recovery()` with an appropriate logger.