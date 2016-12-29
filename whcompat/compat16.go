// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// This file can go once everything uses go1.7 context semantics

// +build !go1.7

package whcompat

import (
	"net/http"
	"net/url"
	"runtime"
	"sync"

	"gopkg.in/webhelp.v1/whroute"
	"golang.org/x/net/context"
)

type reqInfo struct {
	ctx context.Context
}

var (
	reqCtxMappingsMtx sync.Mutex
	reqCtxMappings    = map[*url.URL]reqInfo{}
)

// Context is a light wrapper around the behavior of Go 1.7's
// (*http.Request).Context method, except this version works with earlier Go
// releases, too. In Go 1.7 and on, this simply calls r.Context(). See the
// note for WithContext for how this works on previous Go releases.
// If building with the appengine tag, when needed, fresh contexts will be
// generated with appengine.NewContext().
func Context(r *http.Request) context.Context {
	reqCtxMappingsMtx.Lock()
	info, ok := reqCtxMappings[r.URL]
	if ok {
		reqCtxMappingsMtx.Unlock()
		return info.ctx
	}

	ctx := new16Context(r)
	bindContextAndUnlock(r, ctx)

	return ctx
}

func bindContextAndUnlock(r *http.Request, ctx context.Context) {
	reqCtxMappings[r.URL] = reqInfo{ctx: ctx}
	reqCtxMappingsMtx.Unlock()

	runtime.SetFinalizer(r, func(r *http.Request) {
		reqCtxMappingsMtx.Lock()
		delete(reqCtxMappings, r.URL)
		reqCtxMappingsMtx.Unlock()
	})
}

func copyReqAndURL(r *http.Request) (c *http.Request) {
	c = &http.Request{}
	*c = *r
	c.URL = &url.URL{}
	*(c.URL) = *(r.URL)
	return c
}

// WithContext is a light wrapper around the behavior of Go 1.7's
// (*http.Request).WithContext method, except this version works with earlier
// Go releases, too. IMPORTANT CAVEAT: to get this to work for Go 1.6 and
// earlier, a few tricks are pulled, such as expecting the returned r.URL to
// never change what object it points to, and a finalizer is set on the
// returned request.
func WithContext(r *http.Request, ctx context.Context) *http.Request {
	if ctx == nil {
		panic("nil ctx")
	}
	r = copyReqAndURL(r)
	reqCtxMappingsMtx.Lock()
	bindContextAndUnlock(r, ctx)
	return r
}

// DoneNotify cancels request contexts when the http.Handler returns in Go
// releases prior to Go 1.7. In Go 1.7 and forward, this is a no-op.
// You get this behavior for free if you use whlog.ListenAndServe.
func DoneNotify(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(Context(r))
			defer cancel()
			h.ServeHTTP(w, WithContext(r, ctx))
		})
}
