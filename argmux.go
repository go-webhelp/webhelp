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
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		var arg string
		arg, r.URL.Path = Shift(r.URL.Path)
		return h.HandleHTTP(context.WithValue(ctx, a, arg), w, r)
	})
}

// ShiftIf is like Shift but will use the second handler if there's no argument
// found.
func (a StringArgMux) ShiftIf(found Handler, notfound Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		arg, newpath := Shift(r.URL.Path)
		if arg == "" {
			return notfound.HandleHTTP(ctx, w, r)
		}
		r.URL.Path = newpath
		return found.HandleHTTP(context.WithValue(ctx, a, arg), w, r)
	})
}

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

// Shift takes a Handler and returns a new Handler that does additional request
// processing. When an incoming request is processed, the new Handler pulls the
// next path element off of the incoming request path and puts the value in the
// current Context. It then passes processing off to the wrapped Handler. It
// responds with a 404 if no numeric value is found.
func (a IntArgMux) Shift(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		str, newpath := Shift(r.URL.Path)
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return ErrNotFound.New("resource: %#v", str)
		}
		r.URL.Path = newpath
		return h.HandleHTTP(context.WithValue(ctx, a, val), w, r)
	})
}

// ShiftIf is like Shift but will use the second handler if there's no numeric
// argument found.
func (a IntArgMux) ShiftIf(found Handler, notfound Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w ResponseWriter,
		r *http.Request) error {
		str, newpath := Shift(r.URL.Path)
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return notfound.HandleHTTP(ctx, w, r)
		}
		r.URL.Path = newpath
		return found.HandleHTTP(context.WithValue(ctx, a, val), w, r)
	})
}

// Get returns a stored value for the Arg from the Context, or 0 if no value
// was found (which won't be the case if a higher-level handler was this
// argmux)
func (a IntArgMux) Get(ctx context.Context) (val int64) {
	if val, ok := ctx.Value(a).(int64); ok {
		return val
	}
	return 0
}
