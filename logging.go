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

type loggingHandler struct {
	h http.Handler
}

// LoggingHandler takes a Handler and makes it log requests.
func LoggingHandler(h http.Handler) http.Handler {
	return loggingHandler{h: h}
}

func (lh loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, requestURI := r.Method, r.RequestURI
	rw := wrapResponseWriter(w)

	logger.Infof("%s %s", method, requestURI)
	var panicked bool
	func() {
		defer func() {
			rec := recover()
			if rec != nil {
				panicked = true
				if rec != stopHandling {
					panic(rec)
				}
			}
		}()
		lh.h.ServeHTTP(w, r)
	}()

	if !rw.WroteHeader() {
		if panicked {
			rw.WriteHeader(http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
	}

	code := rw.StatusCode()

	if code >= 200 && code < 300 {
		logger.Noticef("%s %s: %d", method, requestURI, code)
	} else {
		logger.Errorf("%s %s: %d", method, requestURI, code)
	}
}

func (lh loggingHandler) Routes(cb func(method, path string,
	annotations []string)) {
	Routes(lh.h, cb)
}

var _ RouteLister = loggingHandler{}
var _ http.Handler = loggingHandler{}
