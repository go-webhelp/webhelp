// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whroute provides utilities to implement route listing, whereby
// http.Handlers that opt in can list what routes they understand.
package whroute

import (
	"fmt"
	"io"
	"net/http"
	"sort"
)

const (
	// AllMethods should be returned from a whroute.Lister when all methods are
	// successfully handled.
	AllMethods = "ALL"

	// AllPaths should be returned from a whroute.Lister when all paths are
	// successfully handled.
	AllPaths = "[/<*>]"
)

// Lister is an interface handlers can implement if they want the Routes
// method to work. All http.Handlers in the webhelp package implement Routes.
type Lister interface {
	Routes(cb func(method, path string, annotations map[string]string))
}

// Routes will call cb with all routes known to h.
func Routes(h http.Handler,
	cb func(method, path string, annotations map[string]string)) {
	if rl, ok := h.(Lister); ok {
		rl.Routes(cb)
	} else {
		cb(AllMethods, AllPaths, nil)
	}
}

// PrintRoutes will write all routes of h to out, using the Routes method.
func PrintRoutes(out io.Writer, h http.Handler) (err error) {
	Routes(h, func(method, path string, annotations map[string]string) {
		if err != nil {
			return
		}
		if host, ok := annotations["Host"]; ok {
			_, err = fmt.Fprintf(out, "%s\t%s%s\n", method, host, path)
		} else {
			_, err = fmt.Fprintf(out, "%s\t%s\n", method, path)
		}
		annotationKeys := make([]string, 0, len(annotations))
		for key := range annotations {
			annotationKeys = append(annotationKeys, key)
		}
		sort.Strings(annotationKeys)
		for _, key := range annotationKeys {
			if err != nil {
				return
			}
			if key == "Host" {
				continue
			}
			_, err = fmt.Fprintf(out, " %s: %s\n", key, annotations[key])
		}
	})
	return err
}

type routeHandlerFunc struct {
	routes http.Handler
	fn     func(http.ResponseWriter, *http.Request)
}

// HandlerFunc advertises the routes from routes, but serves content using
// fn.
func HandlerFunc(routes http.Handler,
	fn func(http.ResponseWriter, *http.Request)) http.Handler {
	return routeHandlerFunc{
		routes: routes,
		fn:     fn}
}

func (rhf routeHandlerFunc) Routes(
	cb func(method, path string, annotations map[string]string)) {
	Routes(rhf.routes, cb)
}

func (rhf routeHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rhf.fn(w, r)
}
