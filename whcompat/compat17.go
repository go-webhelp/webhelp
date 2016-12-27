// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// This file can go once everything uses go1.7 context semantics

// +build go1.7

package whcompat

import (
	"context"
	"net/http"
)

// Context is a light wrapper around the behavior of Go 1.7's
// (*http.Request).Context method, except this version works with earlier Go
// releases, too. In Go 1.7 and on, this simply calls r.Context(). See the
// note for WithContext for how this works on previous Go releases.
// If building with the appengine tag, when needed, fresh contexts will be
// generated with appengine.NewContext().
func Context(r *http.Request) context.Context {
	return r.Context()
}

// WithContext is a light wrapper around the behavior of Go 1.7's
// (*http.Request).WithContext method, except this version works with earlier
// Go releases, too. IMPORTANT CAVEAT: to get this to work for Go 1.6 and
// earlier, a few tricks are pulled, such as expecting the returned r.URL to
// never change what object it points to, and a finalizer is set on the
// returned request.
func WithContext(r *http.Request, ctx context.Context) *http.Request {
	return r.WithContext(ctx)
}

// DoneNotify cancels request contexts when the http.Handler returns in Go
// releases prior to Go 1.7. In Go 1.7 and forward, this is a no-op.
// You get this behavior for free if you use whlog.ListenAndServe.
func DoneNotify(h http.Handler) http.Handler {
	return h
}
