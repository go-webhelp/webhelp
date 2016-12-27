// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whlog provides functionality to log incoming requests and results.
package whlog

import (
	"net/http"

	"github.com/jtolds/webhelp/whmon"
	"github.com/jtolds/webhelp/whroute"
	"github.com/spacemonkeygo/spacelog"
)

var (
	logger = spacelog.GetLogger()
)

// LogRequests takes a Handler and makes it log requests. LogRequests uses
// whmon's ResponseWriter to keep track of activity. whfatal.Catch should be
// placed *inside* if applicable.
func LogRequests(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			method, requestURI := r.Method, r.RequestURI
			rw := whmon.WrapResponseWriter(w)

			logger.Infof("%s %s", method, requestURI)

			defer func() {
				rec := recover()
				if rec != nil {
					logger.Critf("Panic: %v", rec)
					panic(rec)
				}
			}()
			h.ServeHTTP(rw, r)

			if !rw.WroteHeader() {
				rw.WriteHeader(http.StatusOK)
			}

			code := rw.StatusCode()

			level := spacelog.Error
			if code >= 200 && code < 300 {
				level = spacelog.Notice
			}

			logger.Logf(level, `%s %#v %d %d %d`, method, requestURI, code,
				r.ContentLength, rw.Written())
		})
}
