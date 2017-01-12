// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whfatal uses panics to make early termination of http.Handlers
// easier. No other webhelp package depends on or uses this one.
package whfatal // import "gopkg.in/webhelp.v1/whfatal"

import (
	"net/http"

	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whmon"
	"gopkg.in/webhelp.v1/whredir"
	"gopkg.in/webhelp.v1/whroute"
)

type fatalBehavior func(w http.ResponseWriter, r *http.Request)

// Catch takes a Handler and returns a new one that works with Fatal,
// whfatal.Redirect, and whfatal.Error. Catch will also catch panics that are
// wherr.HTTPError errors. Catch should be placed *inside* a whlog.LogRequests
// handler, wherr.HandleWith handlers, and a few other handlers. Otherwise,
// the wrapper will be one of the things interrupted by Fatal calls.
func Catch(h http.Handler) http.Handler {
	return whmon.MonitorResponse(whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			rw := w.(whmon.ResponseWriter)
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				behavior, ok := rec.(fatalBehavior)
				if !ok {
					perr, ok := rec.(error)
					if !ok || !wherr.HTTPError.Contains(perr) {
						panic(rec)
					}
					behavior = func(w http.ResponseWriter, r *http.Request) {
						wherr.Handle(w, r, perr)
					}
				}
				if behavior != nil {
					behavior(rw, r)
				}
				if !rw.WroteHeader() {
					rw.WriteHeader(http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(rw, r)
		}))
}

// Redirect is like whredir.Redirect but panics so all additional request
// processing terminates. Implemented with Fatal().
//
// IMPORTANT: must be used with whfatal.Catch, or else the http.ResponseWriter
// won't be able to be obtained. Because this requires whfatal.Catch, if
// you're writing a library intended to be used by others, please avoid this
// and other Fatal* methods. If you are writing a library intended to be used
// by yourself, you should probably avoid these methods anyway.
func Redirect(redirectTo string) {
	Fatal(func(w http.ResponseWriter, r *http.Request) {
		whredir.Redirect(w, r, redirectTo)
	})
}

// Error is like wherr.Handle but panics so that all additional request
// processing terminates. Implemented with Fatal()
//
// IMPORTANT: must be used with whfatal.Catch, or else the http.ResponseWriter
// won't be able to be obtained. Because this requires whfatal.Catch, if
// you're writing a library intended to be used by others, please avoid this
// and other Fatal* methods. If you are writing a library intended to be used
// by yourself, you should probably avoid these methods anyway.
func Error(err error) {
	Fatal(func(w http.ResponseWriter, r *http.Request) {
		wherr.Handle(w, r, err)
	})
}

// Fatal panics in a way that Catch understands to abort all additional
// request processing. Once request processing has been aborted, handler is
// called, if not nil. If handler doesn't write a response, a 500 will
// automatically be returned. whfatal.Error and whfatal.Redirect are
// implemented using this method.
//
// IMPORTANT: must be used with whfatal.Catch, or else the http.ResponseWriter
// won't be able to be obtained. Because this requires whfatal.Catch, if
// you're writing a library intended to be used by others, please avoid this
// and other Fatal* methods. If you are writing a library intended to be used
// by yourself, you should probably avoid these methods anyway.
func Fatal(handler func(w http.ResponseWriter, r *http.Request)) {
	// even if handler == nil, this is NOT the same as panic(nil), so we're okay.
	panic(fatalBehavior(handler))
}
