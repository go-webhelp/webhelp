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

func (d DirMux) Routes(
	cb func(method, path string, annotations map[string]string)) {
	keys := make([]string, 0, len(d))
	for element := range d {
		keys = append(keys, element)
	}
	sort.Strings(keys)
	for _, element := range keys {
		Routes(d[element],
			func(method, path string, annotations map[string]string) {
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

func (m MethodMux) Routes(
	cb func(method, path string, annotations map[string]string)) {
	keys := make([]string, 0, len(m))
	for method := range m {
		keys = append(keys, method)
	}
	sort.Strings(keys)
	for _, method := range keys {
		handler := m[method]
		Routes(handler, func(_, path string, annotations map[string]string) {
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

// HostMux is an http.Handler that chooses a subhandler based on the request
// Host header. The star host ("*") is a default handler.
type HostMux map[string]http.Handler

func (h HostMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := h[r.Host]
	if !ok {
		handler, ok = h["*"]
		if !ok {
			HandleError(w, r, ErrNotFound.New("host: %#v", r.Host))
			return
		}
	}
	handler.ServeHTTP(w, r)
}

func (h HostMux) Routes(
	cb func(method, path string, annotations map[string]string)) {
	keys := make([]string, 0, len(h))
	for element := range h {
		keys = append(keys, element)
	}
	sort.Strings(keys)
	for _, host := range keys {
		Routes(h[host], func(method, path string, annotations map[string]string) {
			cp := make(map[string]string, len(annotations)+1)
			for key, val := range annotations {
				cp[key] = val
			}
			cp["Host"] = host
			cb(method, path, cp)
		})
	}
}

var _ http.Handler = HostMux(nil)
var _ RouteLister = HostMux(nil)

// OverlayMux is essentially a DirMux that you can put in front of another
// http.Handler. If the requested entry isn't in the overlay DirMux, the
// Default will be used. If no Default is specified this works exactly the
// same as a DirMux.
type OverlayMux struct {
	Default http.Handler
	Overlay DirMux
}

func (o OverlayMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, left := Shift(r.URL.Path)
	handler, ok := o.Overlay[dir]
	if !ok {
		if o.Default == nil {
			HandleError(w, r, ErrNotFound.New("resource: %#v", dir))
			return
		}
		o.Default.ServeHTTP(w, r)
		return
	}
	r.URL.Path = left
	handler.ServeHTTP(w, r)
}

func (o OverlayMux) Routes(
	cb func(method, path string, annotations map[string]string)) {
	Routes(o.Overlay, cb)
	if o.Default != nil {
		Routes(o.Default, cb)
	}
}

var _ http.Handler = OverlayMux{}
var _ RouteLister = OverlayMux{}

type redirectHandler string

// RedirectHandler returns an http.Handler that redirects all requests to url.
func RedirectHandler(url string) http.Handler {
	return redirectHandler(url)
}

func (t redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Redirect(w, r, string(t))
}

func (t redirectHandler) Routes(
	cb func(method, path string, annotations map[string]string)) {
	cb(AllMethods, AllPaths, map[string]string{"Redirect": string(t)})
}

var _ http.Handler = redirectHandler("")
var _ RouteLister = redirectHandler("")

// RedirectHandlerFunc is an http.Handler that redirects all requests to the
// returned URL.
type RedirectHandlerFunc func(r *http.Request) string

func (f RedirectHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Redirect(w, r, f(r))
}

func (f RedirectHandlerFunc) Routes(
	cb func(method, path string, annotations map[string]string)) {
	cb(AllMethods, AllPaths, map[string]string{"Redirect": "f(req)"})
}

var _ http.Handler = RedirectHandlerFunc(nil)
var _ RouteLister = RedirectHandlerFunc(nil)

// RequireHTTPS returns a handler that will redirect to the same path but using
// https if https was not already used.
func RequireHTTPS(handler http.Handler) http.Handler {
	return RouteHandlerFunc(handler,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Scheme != "https" {
				u := *r.URL
				u.Scheme = "https"
				Redirect(w, r, u.String())
			} else {
				handler.ServeHTTP(w, r)
			}
		})
}

// RequireHost returns a handler that will redirect to the same path but using
// the given host if the given host was not specifically requested.
func RequireHost(host string, handler http.Handler) http.Handler {
	if host == "*" {
		return handler
	}
	return HostMux{
		host: handler,
		"*": RedirectHandlerFunc(func(r *http.Request) string {
			u := *r.URL
			u.Host = host
			return u.String()
		})}
}
