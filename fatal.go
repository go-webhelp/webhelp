// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"
)

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
			if rec != stopHandling {
				panic(rec)
			}
			if !rw.WroteHeader() {
				rw.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(rw, r)
	})
}

// FatalRedirect is like Redirect but panics so all additional request
// processing terminates. IMPORTANT: must be used with FatalHandler, or else
// the panic prevents the standard library from flushing the response to the
// client.
func FatalRedirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	Redirect(w, r, redirectTo)
	panic(stopHandling)
}

// FatalError is like HandleError but panics so that all additional request
// processing terminates. IMPORTANT: must be used with FatalHandler, or else
// the panic prevents the standard library from flushing the response to the
// client.
func FatalError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err)
	panic(stopHandling)
}
