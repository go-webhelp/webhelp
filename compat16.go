// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// This file can go once everything uses go1.7 context semantics

// +build !go1.7

package webhelp

import (
	"net/http"
	"net/url"
	"runtime"
	"sync"

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
// (*http.Request).Context method, except this version works with Go 1.6 too.
func Context(r *http.Request) context.Context {
	reqCtxMappingsMtx.Lock()
	info, ok := reqCtxMappings[r.URL]
	reqCtxMappingsMtx.Unlock()

	if ok {
		return info.ctx
	}
	return context.Background()
}

// WithContext is a light wrapper around the behavior of Go 1.7's
// (*http.Request).WithContext method, except this version works with Go 1.6
// too. IMPORTANT CAVEAT: to get this to work for Go 1.6, a few tricks are
// pulled, such as expecting the returned r.URL to never change what object it
// points to, and a finalizer is set on the returned request.
func WithContext(r *http.Request, ctx context.Context) *http.Request {
	if ctx == nil {
		panic("nil ctx")
	}

	r2 := &http.Request{}
	*r2 = *r
	r2.URL = &url.URL{}
	*(r2.URL) = *(r.URL)

	reqCtxMappingsMtx.Lock()
	reqCtxMappings[r2.URL] = reqInfo{ctx: ctx}
	reqCtxMappingsMtx.Unlock()

	runtime.SetFinalizer(r2, func(r2 *http.Request) {
		reqCtxMappingsMtx.Lock()
		delete(reqCtxMappings, r2.URL)
		reqCtxMappingsMtx.Unlock()
	})

	return r2
}

// ContextBase is a back-compat handler for Go1.7 context features in Go1.6.
// You'll need to have this at the base of your handler stack. You don't need
// to use this if you're using webhelp.ListenAndServe.
func ContextBase(h http.Handler) http.Handler {
	return RouteHandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(Context(r))
			defer cancel()
			h.ServeHTTP(w, WithContext(r, ctx))
		})
}
