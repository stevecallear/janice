# Janice
[![Build Status](https://travis-ci.org/stevecallear/janice.svg?branch=master)](https://travis-ci.org/stevecallear/janice)
[![codecov](https://codecov.io/gh/stevecallear/janice/branch/master/graph/badge.svg)](https://codecov.io/gh/stevecallear/janice)

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
```