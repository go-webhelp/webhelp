// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"
)

type fatalBehavior func(w http.ResponseWriter, r *http.Request)

// FatalHandler takes a Handler and returns a new one that works with
// Fatal, FatalRedirect, and FatalError. FatalHandler should be placed *inside*
// a LoggingHandler, otherwise logging will be one of the things interrupted
// by Fatal calls.
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
// processing terminates. Implemented with Fatal().
//
// IMPORTANT: must be used with FatalHandler, or else the http.ResponseWriter
// won't be able to be obtained. Because this requires FatalHandler, if
// you're writing a library intended to be used by others, please avoid this
// and other Fatal* methods. If you are writing a library intended to be used
// by yourself, you should probably avoid these methods anyway.
func FatalRedirect(redirectTo string) {
	Fatal(func(w http.ResponseWriter, r *http.Request) {
		Redirect(w, r, redirectTo)
	})
}

// FatalError is like HandleError but panics so that all additional request
// processing terminates. Implemented with Fatal()
//
// IMPORTANT: must be used with FatalHandler, or else the http.ResponseWriter
// won't be able to be obtained. Because this requires FatalHandler, if
// you're writing a library intended to be used by others, please avoid this
// and other Fatal* methods. If you are writing a library intended to be used
// by yourself, you should probably avoid these methods anyway.
func FatalError(err error) {
	Fatal(func(w http.ResponseWriter, r *http.Request) {
		HandleError(w, r, err)
	})
}

// Fatal panics in a way that FatalHandler understands to abort all additional
// request processing. Once request processing has been aborted, handler is
// called. If handler doesn't write a response, a 500 will automatically be
// returned. FatalError and FatalRedirect are implemented using this method.
//
// IMPORTANT: must be used with FatalHandler, or else the http.ResponseWriter
// won't be able to be obtained. Because this requires FatalHandler, if
// you're writing a library intended to be used by others, please avoid this
// and other Fatal* methods. If you are writing a library intended to be used
// by yourself, you should probably avoid these methods anyway.
func Fatal(handler func(w http.ResponseWriter, r *http.Request)) {
	panic(fatalBehavior(handler))
}
