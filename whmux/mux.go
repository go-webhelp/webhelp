// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whmux provides some useful request mux helpers for demultiplexing
// requests to one of a number of handlers.
package whmux // import "gopkg.in/webhelp.v1/whmux"

import (
	"net/http"
	"sort"
	"strings"

	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whroute"
)

// Dir is an http.Handler that mimics a directory. It mutates an incoming
// request's URL.Path to properly namespace handlers. This way a handler can
// assume it has the root of its section. If you want the original URL, use
// req.RequestURI (but don't modify it).
type Dir map[string]http.Handler

// ServeHTTP implements http.handler
func (d Dir) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, left := Shift(r.URL.Path)
	handler, ok := d[dir]
	if !ok {
		wherr.Handle(w, r, wherr.NotFound.New("resource: %#v", dir))
		return
	}
	if left == "" {
		left = "/"
	}
	r.URL.Path = left
	handler.ServeHTTP(w, r)
}

// Routes implements whroute.Lister
func (d Dir) Routes(
	cb func(method, path string, annotations map[string]string)) {
	keys := make([]string, 0, len(d))
	for element := range d {
		keys = append(keys, element)
	}
	sort.Strings(keys)
	for _, element := range keys {
		whroute.Routes(d[element],
			func(method, path string, annotations map[string]string) {
				if element == "" {
					cb(method, "/", annotations)
				} else {
					cb(method, "/"+element+path, annotations)
				}
			})
	}
}

var _ http.Handler = Dir(nil)
var _ whroute.Lister = Dir(nil)

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

// Method is an http.Handler muxer that keys off of the given HTTP request
// method.
type Method map[string]http.Handler

// ServeHTTP implements http.handler
func (m Method) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, found := m[r.Method]; found {
		handler.ServeHTTP(w, r)
		return
	}
	wherr.Handle(w, r,
		wherr.MethodNotAllowed.New("bad method: %#v", r.Method))
}

// Routes implements whroute.Lister
func (m Method) Routes(
	cb func(method, path string, annotations map[string]string)) {
	keys := make([]string, 0, len(m))
	for method := range m {
		keys = append(keys, method)
	}
	sort.Strings(keys)
	for _, method := range keys {
		handler := m[method]
		whroute.Routes(handler,
			func(_, path string, annotations map[string]string) {
				cb(method, path, annotations)
			})
	}
}

var _ http.Handler = Method(nil)
var _ whroute.Lister = Method(nil)

// ExactPath takes an http.Handler that returns a new http.Handler that doesn't
// accept any more path elements and returns a 404 if more are provided.
func ExactPath(h http.Handler) http.Handler {
	return Dir{"": h}
}

// RequireMethod takes an http.Handler and returns a new http.Handler that only
// works with the given HTTP method. If a different method is used, a 405 is
// returned.
func RequireMethod(method string, h http.Handler) http.Handler {
	return Method{method: h}
}

// RequireGet is simply RequireMethod but called with "GET" as the first
// argument.
func RequireGet(h http.Handler) http.Handler {
	return RequireMethod("GET", h)
}

// Exact is simply RequireGet and ExactPath called together.
func Exact(h http.Handler) http.Handler {
	return RequireGet(ExactPath(h))
}

// Host is an http.Handler that chooses a subhandler based on the request
// Host header. The star host ("*") is a default handler.
type Host map[string]http.Handler

// ServeHTTP implements http.handler
func (h Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := h[r.Host]
	if !ok {
		handler, ok = h["*"]
		if !ok {
			wherr.Handle(w, r, wherr.NotFound.New("host: %#v", r.Host))
			return
		}
	}
	handler.ServeHTTP(w, r)
}

// Routes implements whroute.Lister
func (h Host) Routes(
	cb func(method, path string, annotations map[string]string)) {
	keys := make([]string, 0, len(h))
	for element := range h {
		keys = append(keys, element)
	}
	sort.Strings(keys)
	for _, host := range keys {
		whroute.Routes(h[host],
			func(method, path string, annotations map[string]string) {
				cp := make(map[string]string, len(annotations)+1)
				for key, val := range annotations {
					cp[key] = val
				}
				cp["Host"] = host
				cb(method, path, cp)
			})
	}
}

var _ http.Handler = Host(nil)
var _ whroute.Lister = Host(nil)

// Overlay is essentially a Dir that you can put in front of another
// http.Handler. If the requested entry isn't in the overlay Dir, the
// Default will be used. If no Default is specified this works exactly the
// same as a Dir.
type Overlay struct {
	Default http.Handler
	Overlay Dir
}

// ServeHTTP implements http.handler
func (o Overlay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, left := Shift(r.URL.Path)
	handler, ok := o.Overlay[dir]
	if !ok {
		if o.Default == nil {
			wherr.Handle(w, r, wherr.NotFound.New("resource: %#v", dir))
			return
		}
		o.Default.ServeHTTP(w, r)
		return
	}
	r.URL.Path = left
	handler.ServeHTTP(w, r)
}

// Routes implements whroute.Lister
func (o Overlay) Routes(
	cb func(method, path string, annotations map[string]string)) {
	whroute.Routes(o.Overlay, cb)
	if o.Default != nil {
		whroute.Routes(o.Default, cb)
	}
}

var _ http.Handler = Overlay{}
var _ whroute.Lister = Overlay{}
