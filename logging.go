// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"

	"github.com/spacemonkeygo/spacelog"
)

var (
	logger = spacelog.GetLogger()
)

// LoggingHandler takes a Handler and makes it log requests. FatalHandlers
// should be placed *inside* LoggingHandlers if applicable.
func LoggingHandler(h http.Handler) http.Handler {
	return RouteHandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		method, requestURI := r.Method, r.RequestURI
		rw := wrapResponseWriter(w)

		logger.Infof("%s %s", method, requestURI)
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
