// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"
	"strconv"
	"sync/atomic"

	"golang.org/x/net/context"
)

// StringArgMux is a way to pull off arbitrary path elements from an incoming
// URL. You'll need to create one with NewStringArgMux.
type StringArgMux int64

var argMuxCounter int64

func NewStringArgMux() StringArgMux {
	return StringArgMux(atomic.AddInt64(&argMuxCounter, 1))
}

// Shift takes a Handler and returns a new Handler that does additional request
// processing. When an incoming request is processed, the new Handler pulls the
// next path element off of the incoming request path and puts the value in the
// current Context. It then passes processing off to the wrapped Handler.
// The value will be an empty string if no argument is found.
func (a StringArgMux) Shift(h Handler) Handler {
	return a.OptShift(h, h)
}

type stringOptShift struct {
	a               StringArgMux
	found, notfound Handler
}

// OptShift is like Shift but the first handler is used only if there's no
// argument found, and the second handler is used if there is.
func (a StringArgMux) OptShift(notfound Handler, found Handler) Handler {
	return stringOptShift{a: a, found: found, notfound: notfound}
}

func (ssi stringOptShift) HandleHTTP(ctx context.Context, w ResponseWriter,
	r *http.Request) error {
	arg, newpath := Shift(r.URL.Path)
	if arg == "" {
		return ssi.notfound.HandleHTTP(ctx, w, r)
	}
	r.URL.Path = newpath
	return ssi.found.HandleHTTP(context.WithValue(ctx, ssi.a, arg), w, r)
}

func (ssi stringOptShift) Routes(cb func(string, string, []string)) {
	Routes(ssi.found, func(method, path string, annotations []string) {
		cb(method, "/<string>"+path, annotations)
	})
	Routes(ssi.notfound, cb)
}

var _ Handler = stringOptShift{}
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

func (notFoundHandler) HandleHTTP(ctx context.Context, w ResponseWriter,
	r *http.Request) error {
	return ErrNotFound.New("resource: %#v", r.URL.Path)
}

func (notFoundHandler) Routes(cb func(string, string, []string)) {}

var _ Handler = notFoundHandler{}
var _ RouteLister = notFoundHandler{}

// Shift takes a Handler and returns a new Handler that does additional request
// processing. When an incoming request is processed, the new Handler pulls the
// next path element off of the incoming request path and puts the value in the
// current Context. It then passes processing off to the wrapped Handler. It
// responds with a 404 if no numeric value is found.
func (a IntArgMux) Shift(h Handler) Handler {
	return a.OptShift(notFoundHandler{}, h)
}

type intOptShift struct {
	a               IntArgMux
	found, notfound Handler
}

// OptShift is like Shift but will only use the second handler if there's no
// numeric argument found and the first handler otherwise
func (a IntArgMux) OptShift(notfound Handler, found Handler) Handler {
	return intOptShift{a: a, found: found, notfound: notfound}
}

func (isi intOptShift) HandleHTTP(ctx context.Context, w ResponseWriter,
	r *http.Request) error {
	str, newpath := Shift(r.URL.Path)
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return isi.notfound.HandleHTTP(ctx, w, r)
	}
	r.URL.Path = newpath
	return isi.found.HandleHTTP(context.WithValue(ctx, isi.a, val), w, r)
}

func (isi intOptShift) Routes(cb func(string, string, []string)) {
	Routes(isi.found, func(method, path string, annotations []string) {
		cb(method, "/<int>"+path, annotations)
	})
	Routes(isi.notfound, cb)
}

var _ Handler = intOptShift{}
var _ RouteLister = intOptShift{}

// Get returns a stored value for the Arg from the Context, or 0 if no value
// was found (which won't be the case if a higher-level handler was this
// argmux)
func (a IntArgMux) Get(ctx context.Context) (val int64) {
	if val, ok := ctx.Value(a).(int64); ok {
		return val
	}
	return 0
}
