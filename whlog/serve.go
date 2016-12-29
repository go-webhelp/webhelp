// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package whlog

import (
	"log"
	"net"
	"net/http"
	"time"

	"gopkg.in/webhelp.v1/whcompat"
)

const (
	keepAlivePeriod = 3 * time.Minute
)

func serve(l net.Listener, handler http.Handler) error {
	if tcp_l, ok := l.(*net.TCPListener); ok {
		l = tcpKeepAliveListener{TCPListener: tcp_l}
	}
	return (&http.Server{Handler: handler}).Serve(l)
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(keepAlivePeriod)
	return tc, nil
}

// ListenAndServe creates a TCP listener prior to calling Serve. It also logs
// the address it listens on, and wraps given handlers in whcompat.DoneNotify.
// Like the standard library, it sets TCP keepalive semantics on.
func ListenAndServe(addr string, handler http.Handler) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", l.Addr())
	return serve(l, whcompat.DoneNotify(handler))
}
