// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"fmt"
	"io"
	"net/http"
)

const (
	AllMethods = "ALL"
	AllPaths   = "[/<*>]"
)

// RouteLister is an interface handlers can implement if they want Routes to
// work.
type RouteLister interface {
	Routes(cb func(method, path string, annotations []string))
}

// Routes will call cb with all routes known to h.
func Routes(h http.Handler,
	cb func(method, path string, annotations []string)) {
	if rl, ok := h.(RouteLister); ok {
		rl.Routes(cb)
	} else {
		cb(AllMethods, AllPaths, nil)
	}
}

// PrintRoutes will write all routes of h to out.
func PrintRoutes(out io.Writer, h http.Handler) (err error) {
	Routes(h, func(method, path string, annotations []string) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(out, "%s %s\n", method, path)
		for _, annotation := range annotations {
			if err != nil {
				return
			}
			_, err = fmt.Fprintf(out, " %s\n", annotation)
		}
	})
	return err
}

type routeHandlerFunc struct {
	routes http.Handler
	fn     func(http.ResponseWriter, *http.Request)
}

// RouteHandlerFunc advertises the routes from routes, but serves content using
// fn.
func RouteHandlerFunc(routes http.Handler,
	fn func(http.ResponseWriter, *http.Request)) http.Handler {
	return routeHandlerFunc{
		routes: routes,
		fn:     fn}
}

func (rhf routeHandlerFunc) Routes(
	cb func(method, path string, annotations []string)) {
	Routes(rhf.routes, cb)
}

func (rhf routeHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rhf.fn(w, r)
}
