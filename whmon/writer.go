// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whmon provides a means to monitor various aspects of how the request
// and response is going.
package whmon // import "gopkg.in/webhelp.v1/whmon"

import (
	"net/http"

	"gopkg.in/webhelp.v1/whroute"
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

type ResponseWriter interface {
	// Header, Write, and WriteHeader are exactly like http.ResponseWriter
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)

	// WroteHeader returns true if the Header was sent out. Note that this can
	// happen if only Write is called.
	WroteHeader() bool
	// StatusCode returns the HTTP status code the Header sent.
	StatusCode() int
	// Written returns the total amount of bytes successfully passed through the
	// Write call. This does not include the header.
	Written() int64
}

// MonitorResponse wraps all incoming http.ResponseWriters with a
// monitored ResponseWriter that keeps track of additional status information
// about the outgoing response. It preserves whether or not the passed in
// response writer is an http.Flusher, http.CloseNotifier, or an http.Hijacker.
// whlog.LogRequests and whfatal.Catch also do this for you.
func MonitorResponse(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(wrapResponseWriter(w), r)
		})
}

// wrapResponseWriter will wrap an http.ResponseWriter with the instrumentation
// to turn it into a whmon.ResponseWriter. An http.ResponseWriter must be
// turned into a whmon.ResponseWriter before being used. It's much better
// to use MonitorResponse instead of WrapResponseWriter.
func wrapResponseWriter(w http.ResponseWriter) ResponseWriter {
	if ww, ok := w.(ResponseWriter); ok {
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
					ResponseWriter
					http.Flusher
					http.CloseNotifier
					http.Hijacker
				}{
					ResponseWriter: rw,
					Flusher:        fw,
					CloseNotifier:  cnw,
					Hijacker:       hw,
				}
			} else {
				return struct {
					ResponseWriter
					http.Flusher
					http.CloseNotifier
				}{
					ResponseWriter: rw,
					Flusher:        fw,
					CloseNotifier:  cnw,
				}
			}
		} else {
			if hok {
				return struct {
					ResponseWriter
					http.Flusher
					http.Hijacker
				}{
					ResponseWriter: rw,
					Flusher:        fw,
					Hijacker:       hw,
				}
			} else {
				return struct {
					ResponseWriter
					http.Flusher
				}{
					ResponseWriter: rw,
					Flusher:        fw,
				}
			}
		}
	} else {
		if cnok {
			if hok {
				return struct {
					ResponseWriter
					http.CloseNotifier
					http.Hijacker
				}{
					ResponseWriter: rw,
					CloseNotifier:  cnw,
					Hijacker:       hw,
				}
			} else {
				return struct {
					ResponseWriter
					http.CloseNotifier
				}{
					ResponseWriter: rw,
					CloseNotifier:  cnw,
				}
			}
		} else {
			if hok {
				return struct {
					ResponseWriter
					http.Hijacker
				}{
					ResponseWriter: rw,
					Hijacker:       hw,
				}
			} else {
				return rw
			}
		}
	}
}
