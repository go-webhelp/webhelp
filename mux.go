// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"
	"sort"
	"strings"
)

// DirMux is an http.Handler that mimics a directory. It mutates an incoming
// request's URL.Path to properly namespace handlers. This way a handler can
// assume it has the root of its section. If you want the original URL, use
// req.RequestURI (but don't modify it).
type DirMux map[string]http.Handler

func (d DirMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, left := Shift(r.URL.Path)
	handler, ok := d[dir]
	if !ok {
		HandleError(w, r, ErrNotFound.New("resource: %#v", dir))
		return
	}
	r.URL.Path = left
	handler.ServeHTTP(w, r)
}

func (d DirMux) Routes(cb func(method, path string, annotations []string)) {
	keys := make([]string, 0, len(d))
	for element := range d {
		keys = append(keys, element)
	}
	sort.Strings(keys)
	for _, element := range keys {
		Routes(d[element], func(method, path string, annotations []string) {
			if element == "" {
				cb(method, "/", annotations)
			} else {
				cb(method, "/"+element+path, annotations)
			}
		})
	}
}

var _ http.Handler = DirMux(nil)
var _ RouteLister = DirMux(nil)

// Shift pulls the first directory out of the path and returns the remainder.
func Shift(path string) (dir, left string) {
	// slice off the first "/"s if they exists
	path = strings.TrimLeft(path, "/")

	if len(path) == 0 {
		return "", ""
	}

	// find the first '/' after the initial one
	split := strings.Index(path, "/")
	if split == -1 {
		return path, ""
	}
	return path[:split], path[split:]
}

// MethodMux is an http.Handler muxer that keys off of the given HTTP request
// method
type MethodMux map[string]http.Handler

func (m MethodMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, found := m[r.Method]; found {
		handler.ServeHTTP(w, r)
		return
	}
	HandleError(w, r, ErrMethodNotAllowed.New("bad method: %#v", r.Method))
}

func (m MethodMux) Routes(cb func(method, path string, annotations []string)) {
	keys := make([]string, 0, len(m))
	for method := range m {
		keys = append(keys, method)
	}
	sort.Strings(keys)
	for _, method := range keys {
		handler := m[method]
		Routes(handler, func(_, path string, annotations []string) {
			cb(method, path, annotations)
		})
	}
}

var _ http.Handler = MethodMux(nil)
var _ RouteLister = MethodMux(nil)

// ExactPath takes an http.Handler that returns a new http.Handler that doesn't
// accept any more path elements
func ExactPath(h http.Handler) http.Handler {
	return DirMux{"": h}
}

// ExactMethod takes an http.Handler and returns a new http.Handler that only
// works with the given HTTP method.
func ExactMethod(method string, h http.Handler) http.Handler {
	return MethodMux{method: h}
}

// ExactGet is simply ExactMethod but called with "GET" as the first argument.
func ExactGet(h http.Handler) http.Handler {
	return ExactMethod("GET", h)
}

// Exact is simply ExactGet and ExactPath called together.
func Exact(h http.Handler) http.Handler {
	return ExactGet(ExactPath(h))
}

// OverlayMux is essentially a DirMux that you can put in front of another
// http.Handler. If the requested entry isn't in the overlay DirMux, the
// Fallback will be used. If no Fallback is specified this works exactly the
// same as a DirMux.
type OverlayMux struct {
	Fallback http.Handler
	Overlay  DirMux
}

func (o OverlayMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, left := Shift(r.URL.Path)
	handler, ok := o.Overlay[dir]
	if !ok {
		if o.Fallback == nil {
			HandleError(w, r, ErrNotFound.New("resource: %#v", dir))
			return
		}
		o.Fallback.ServeHTTP(w, r)
		return
	}
	r.URL.Path = left
	handler.ServeHTTP(w, r)
}

func (o OverlayMux) Routes(cb func(method, path string, annotations []string)) {
	Routes(o.Overlay, cb)
	if o.Fallback != nil {
		Routes(o.Fallback, cb)
	}
}

var _ http.Handler = OverlayMux{}
var _ RouteLister = OverlayMux{}

// RedirectHandler(url) is an http.Handler that redirects all requests to url.
type RedirectHandler string

func (t RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Redirect(w, r, string(t))
}

func (target RedirectHandler) Routes(
	cb func(method, path string, annotations []string)) {
	cb(AllMethods, AllPaths, []string{"-> " + string(target)})
}

var _ http.Handler = RedirectHandler("")
var _ RouteLister = RedirectHandler("")

// RedirectHandlerFunc is an http.Handler that redirects all requests to the
// returned URL.
type RedirectHandlerFunc func(r *http.Request) string

func (f RedirectHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Redirect(w, r, f(r))
}

func (f RedirectHandlerFunc) Routes(
	cb func(method, path string, annotations []string)) {
	cb(AllMethods, AllPaths, []string{"-> f(req)"})
}

var _ http.Handler = RedirectHandlerFunc(nil)
var _ RouteLister = RedirectHandlerFunc(nil)
