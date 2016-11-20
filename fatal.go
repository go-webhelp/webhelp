// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"
)

type fatalBehavior func(w http.ResponseWriter, r *http.Request)

// FatalHandler takes a Handler and returns a new one that works with
// FatalRedirect and FatalError.
func FatalHandler(h http.Handler) http.Handler {
	return RouteHandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		rw := wrapResponseWriter(w)
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			behavior, ok := rec.(fatalBehavior)
			if !ok {
				panic(rec)
			}
			behavior(rw, r)
			if !rw.WroteHeader() {
				rw.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(rw, r)
	})
}

// FatalRedirect is like Redirect but panics so all additional request
// processing terminates. IMPORTANT: must be used with FatalHandler, or else
// the http.ResponseWriter won't be able to be obtained.
func FatalRedirect(redirectTo string) {
	panic(fatalBehavior(func(w http.ResponseWriter, r *http.Request) {
		Redirect(w, r, redirectTo)
	}))
}

// FatalError is like HandleError but panics so that all additional request
// processing terminates. IMPORTANT: must be used with FatalHandler, or else
// the http.ResponseWriter won't be able to be obtained.
func FatalError(err error) {
	panic(fatalBehavior(func(w http.ResponseWriter, r *http.Request) {
		HandleError(w, r, err)
	}))
}
