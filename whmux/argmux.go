// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package whmux

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"gopkg.in/webhelp.v1"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whroute"
)

// StringArg is a way to pull off arbitrary path elements from an incoming
// URL. You'll need to create one with NewStringArg.
type StringArg webhelp.ContextKey

func NewStringArg() StringArg {
	return StringArg(webhelp.GenSym())
}

// Shift takes an http.Handler and returns a new http.Handler that does
// additional request processing. When an incoming request is processed, the
// new http.Handler pulls the next path element off of the incoming request
// path and puts the value in the current Context. It then passes processing
// off to the wrapped http.Handler. The value will be an empty string if no
// argument is found.
func (a StringArg) Shift(h http.Handler) http.Handler {
	return a.ShiftOpt(h, notFoundHandler{})
}

type stringOptShift struct {
	a               StringArg
	found, notfound http.Handler
}

// ShiftOpt is like Shift but the first handler is used only if there's an
// argument found and the second handler is used if there isn't.
func (a StringArg) ShiftOpt(found, notfound http.Handler) http.Handler {
	return stringOptShift{a: a, found: found, notfound: notfound}
}

func (ssi stringOptShift) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	arg, newpath := Shift(r.URL.Path)
	if arg == "" {
		ssi.notfound.ServeHTTP(w, r)
		return
	}
	r.URL.Path = newpath
	ctx := context.WithValue(whcompat.Context(r), ssi.a, arg)
	ssi.found.ServeHTTP(w, whcompat.WithContext(r, ctx))
}

func (ssi stringOptShift) Routes(cb func(string, string, map[string]string)) {
	whroute.Routes(ssi.found,
		func(method, path string, annotations map[string]string) {
			cb(method, "/<string>"+path, annotations)
		})
	whroute.Routes(ssi.notfound,
		func(method, path string, annotations map[string]string) {
			switch path {
			case whroute.AllPaths, "/":
				cb(method, "/", annotations)
			}
		})
}

var _ http.Handler = stringOptShift{}
var _ whroute.Lister = stringOptShift{}

// Get returns a stored value for the Arg from the Context, or "" if no value
// was found (which won't be the case if a higher-level handler was this
// arg)
func (a StringArg) Get(ctx context.Context) (val string) {
	if val, ok := ctx.Value(a).(string); ok {
		return val
	}
	return ""
}

// IntArg is a way to pull off numeric path elements from an incoming
// URL. You'll need to create one with NewIntArg.
type IntArg webhelp.ContextKey

func NewIntArg() IntArg {
	return IntArg(webhelp.GenSym())
}

type notFoundHandler struct{}

func (notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wherr.Handle(w, r, wherr.NotFound.New("resource: %#v", r.URL.Path))
}

func (notFoundHandler) Routes(cb func(string, string, map[string]string)) {}

var _ http.Handler = notFoundHandler{}
var _ whroute.Lister = notFoundHandler{}

// Shift takes an http.Handler and returns a new http.Handler that does
// additional request processing. When an incoming request is processed, the
// new http.Handler pulls the next path element off of the incoming request
// path and puts the value in the current Context. It then passes processing
// off to the wrapped http.Handler. It responds with a 404 if no numeric value
// is found.
func (a IntArg) Shift(h http.Handler) http.Handler {
	return a.ShiftOpt(h, notFoundHandler{})
}

type intOptShift struct {
	a               IntArg
	found, notfound http.Handler
}

// ShiftOpt is like Shift but will only use the first handler if there's a
// numeric argument found and the second handler otherwise.
func (a IntArg) ShiftOpt(found, notfound http.Handler) http.Handler {
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
	ctx := context.WithValue(whcompat.Context(r), isi.a, val)
	isi.found.ServeHTTP(w, whcompat.WithContext(r, ctx))
}

func (isi intOptShift) Routes(cb func(string, string, map[string]string)) {
	whroute.Routes(isi.found, func(method, path string,
		annotations map[string]string) {
		cb(method, "/<int>"+path, annotations)
	})
	whroute.Routes(isi.notfound, cb)
}

var _ http.Handler = intOptShift{}
var _ whroute.Lister = intOptShift{}

// Get returns a stored value for the Arg from the Context and ok = true if
// found, or ok = false if no value was found (which won't be the case if a
// higher-level handler was this arg)
func (a IntArg) Get(ctx context.Context) (val int64, ok bool) {
	if val, ok := ctx.Value(a).(int64); ok {
		return val, true
	}
	return 0, false
}

// MustGet is like Get but panics in cases when ok would be false. If used with
// whfatal.Catch, will return a 404 to the user.
func (a IntArg) MustGet(ctx context.Context) (val int64) {
	val, ok := ctx.Value(a).(int64)
	if !ok {
		panic(wherr.NotFound.New("Required argument missing"))
	}
	return val
}
