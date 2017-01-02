// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whlog provides functionality to log incoming requests and results.
package whlog // import "gopkg.in/webhelp.v1/whlog"

import (
	"log"
	"net/http"
	"time"

	"gopkg.in/webhelp.v1/whmon"
	"gopkg.in/webhelp.v1/whroute"
)

type Loggerf func(format string, arg ...interface{})

var (
	Default Loggerf = log.Printf
)

// LogResponses takes a Handler and makes it log responses. LogResponses uses
// whmon's ResponseWriter to keep track of activity. whfatal.Catch should be
// placed *inside* if applicable. whlog.Default makes a good default logger.
func LogResponses(logger Loggerf, h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			method, requestURI := r.Method, r.RequestURI
			rw := whmon.WrapResponseWriter(w)
			start := time.Now()

			defer func() {
				rec := recover()
				if rec != nil {
					log.Printf("Panic: %v", rec)
					panic(rec)
				}
			}()
			h.ServeHTTP(rw, r)

			if !rw.WroteHeader() {
				rw.WriteHeader(http.StatusOK)
			}

			code := rw.StatusCode()

			logger(`%s %#v %d %d %d %v`, method, requestURI, code,
				r.ContentLength, rw.Written(), time.Since(start))
		})
}

// LogRequests takes a Handler and makes it log requests (prior to request
// handling). whlog.Default makes a good default logger.
func LogRequests(logger Loggerf, h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			logger(`%s %#v %d`, r.Method, r.RequestURI, r.ContentLength)
			h.ServeHTTP(w, r)
		})
}
