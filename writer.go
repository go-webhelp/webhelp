// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"
)

type rWriter struct {
	w           http.ResponseWriter
	wroteHeader bool
	sc          int
	written     int64
}

var _ http.ResponseWriter = (*rWriter)(nil)

func (r *rWriter) Header() http.Header { return r.w.Header() }

func (r *rWriter) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.sc = 200
	}
	n, err := r.w.Write(p)
	r.written += int64(n)
	return n, err
}

func (r *rWriter) WriteHeader(sc int) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.sc = sc
	}
	r.w.WriteHeader(sc)
}

func (r *rWriter) StatusCode() int {
	if !r.wroteHeader {
		return 200
	}
	return r.sc
}

func (r *rWriter) WroteHeader() bool { return r.wroteHeader }
func (r *rWriter) Written() int64    { return r.written }

type MonitoredResponseWriter interface {
	// Header, Write, and WriteHeader are exactly like http.ResponseWriter
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)

	// WroteHeader returns true if the Header was sent out. Note that this can
	// happen if only Write is called.
	WroteHeader() bool
	// StatusCode returns the HTTP status code the Header sent, if applicable
	StatusCode() int
	// Written returns the total amount of bytes successfully passed through the
	// Write call.
	Written() int64
}

// MonitorResponse wraps all incoming http.ResponseWriters with a
// MonitoredResponseWriter that keeps track of additional status information
// about the outgoing response. It preserves whether or not the passed in
// response writer is an http.Flusher, http.CloseNotifier, or an http.Hijacker.
// LoggingHandler also does this for you.
func MonitorResponse(h http.Handler) http.Handler {
	return RouteHandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(wrapResponseWriter(w), r)
	})
}

// wrapResponseWriter's goal is to make a webhelp.ResponseWriter that has the
// same optional methods as the wrapped http.ResponseWriter
// (Flush, CloseNotify, Hijack). this ends up being SUPER MESSY
func wrapResponseWriter(w http.ResponseWriter) MonitoredResponseWriter {
	if ww, ok := w.(MonitoredResponseWriter); ok {
		// don't do it if we already have the methods we need
		return ww
	}
	fw, fok := w.(http.Flusher)
	cnw, cnok := w.(http.CloseNotifier)
	hw, hok := w.(http.Hijacker)
	rw := &rWriter{w: w}
	if fok {
		if cnok {
			if hok {
				return struct {
					MonitoredResponseWriter
					http.Flusher
					http.CloseNotifier
					http.Hijacker
				}{
					MonitoredResponseWriter: rw,
					Flusher:                 fw,
					CloseNotifier:           cnw,
					Hijacker:                hw,
				}
			} else {
				return struct {
					MonitoredResponseWriter
					http.Flusher
					http.CloseNotifier
				}{
					MonitoredResponseWriter: rw,
					Flusher:                 fw,
					CloseNotifier:           cnw,
				}
			}
		} else {
			if hok {
				return struct {
					MonitoredResponseWriter
					http.Flusher
					http.Hijacker
				}{
					MonitoredResponseWriter: rw,
					Flusher:                 fw,
					Hijacker:                hw,
				}
			} else {
				return struct {
					MonitoredResponseWriter
					http.Flusher
				}{
					MonitoredResponseWriter: rw,
					Flusher:                 fw,
				}
			}
		}
	} else {
		if cnok {
			if hok {
				return struct {
					MonitoredResponseWriter
					http.CloseNotifier
					http.Hijacker
				}{
					MonitoredResponseWriter: rw,
					CloseNotifier:           cnw,
					Hijacker:                hw,
				}
			} else {
				return struct {
					MonitoredResponseWriter
					http.CloseNotifier
				}{
					MonitoredResponseWriter: rw,
					CloseNotifier:           cnw,
				}
			}
		} else {
			if hok {
				return struct {
					MonitoredResponseWriter
					http.Hijacker
				}{
					MonitoredResponseWriter: rw,
					Hijacker:                hw,
				}
			} else {
				return rw
			}
		}
	}
}
