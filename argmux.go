// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
)

// StringArgMux is a way to pull off arbitrary path elements from an incoming
// URL. You'll need to create one with NewStringArgMux.
type StringArgMux int64

var argMuxCounter int64

func NewStringArgMux() StringArgMux {
	return StringArgMux(atomic.AddInt64(&argMuxCounter, 1))
}

// Shift takes an http.Handler and returns a new http.Handler that does
// additional request processing. When an incoming request is processed, the
// new http.Handler pulls the next path element off of the incoming request
// path and puts the value in the current Context. It then passes processing
// off to the wrapped http.Handler. The value will be an empty string if no
// argument is found.
func (a StringArgMux) Shift(h http.Handler) http.Handler {
	return a.ShiftOpt(h, h)
}

type stringOptShift struct {
	a               StringArgMux
	found, notfound http.Handler
}

// ShiftOpt is like Shift but the first handler is used only if there's an
// argument found and the second handler is used if there isn't.
func (a StringArgMux) ShiftOpt(found, notfound http.Handler) http.Handler {
	return stringOptShift{a: a, found: found, notfound: notfound}
}

func (ssi stringOptShift) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	arg, newpath := Shift(r.URL.Path)
	if arg == "" {
		ssi.notfound.ServeHTTP(w, r)
		return
	}
	r.URL.Path = newpath
	ctx := context.WithValue(r.Context(), ssi.a, arg)
	ssi.found.ServeHTTP(w, r.WithContext(ctx))
}

func (ssi stringOptShift) Routes(cb func(string, string, []string)) {
	Routes(ssi.found, func(method, path string, annotations []string) {
		cb(method, "/<string>"+path, annotations)
	})
	Routes(ssi.notfound, cb)
}

var _ http.Handler = stringOptShift{}
var _ RouteLister = stringOptShift{}

// Get returns a stored value for the Arg from the Context, or "" if no value
// was found (which won't be the case if a higher-level handler was this
// argmux)
func (a StringArgMux) Get(ctx context.Context) (val string) {
	if val, ok := ctx.Value(a).(string); ok {
		return val
	}
	return ""
}

// IntArgMux is a way to pull off numeric path elements from an incoming
// URL. You'll need to create one with NewIntArgMux.
type IntArgMux int64

func NewIntArgMux() IntArgMux {
	return IntArgMux(atomic.AddInt64(&argMuxCounter, 1))
}

type notFoundHandler struct{}

func (notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	HandleError(w, r, ErrNotFound.New("resource: %#v", r.URL.Path))
}

func (notFoundHandler) Routes(cb func(string, string, []string)) {}

var _ http.Handler = notFoundHandler{}
var _ RouteLister = notFoundHandler{}

// Shift takes an http.Handler and returns a new http.Handler that does
// additional request processing. When an incoming request is processed, the
// new http.Handler pulls the next path element off of the incoming request
// path and puts the value in the current Context. It then passes processing
// off to the wrapped http.Handler. It responds with a 404 if no numeric value
// is found.
func (a IntArgMux) Shift(h http.Handler) http.Handler {
	return a.ShiftOpt(h, notFoundHandler{})
}

type intOptShift struct {
	a               IntArgMux
	found, notfound http.Handler
}

// ShiftOpt is like Shift but will only use the first handler if there's a
// numeric argument found and the second handler otherwise.
func (a IntArgMux) ShiftOpt(found, notfound http.Handler) http.Handler {
	return intOptShift{a: a, found: found, notfound: notfound}
}

func (isi intOptShift) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	str, newpath := Shift(r.URL.Path)
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		isi.notfound.ServeHTTP(w, r)
		return
	}
	r.URL.Path = newpath
	ctx := context.WithValue(r.Context(), isi.a, val)
	isi.found.ServeHTTP(w, r.WithContext(ctx))
}

func (isi intOptShift) Routes(cb func(string, string, []string)) {
	Routes(isi.found, func(method, path string, annotations []string) {
		cb(method, "/<int>"+path, annotations)
	})
	Routes(isi.notfound, cb)
}

var _ http.Handler = intOptShift{}
var _ RouteLister = intOptShift{}

// Get returns a stored value for the Arg from the Context and ok = true if
// found, or ok = false if no value was found (which won't be the case if a
// higher-level handler was this argmux)
func (a IntArgMux) Get(ctx context.Context) (val int64, ok bool) {
	if val, ok := ctx.Value(a).(int64); ok {
		return val, true
	}
	return 0, false
}

// MustGet is like Get but panics in cases when ok would be false.
func (a IntArgMux) MustGet(ctx context.Context) (val int64) {
	if val, ok := ctx.Value(a).(int64); ok {
		return val
	}
	panic("Required argument missing")
}
