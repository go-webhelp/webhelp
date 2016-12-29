// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// +build !go1.8

package whcompat

import (
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/webhelp.v1/whroute"
)

// CloseNotify causes a handler to have its request.Context() canceled the
// second the client TCP connection goes away by hooking the http.CloseNotifier
// logic into the context. Prior to Go 1.8, this costs an extra goroutine in
// a read loop. Go 1.8 and on, this behavior happens automatically with or
// without this wrapper.
func CloseNotify(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			if cnw, ok := w.(http.CloseNotifier); ok {
				doneChan := make(chan bool)
				defer close(doneChan)

				closeChan := cnw.CloseNotify()
				ctx, cancelFunc := context.WithCancel(Context(r))
				r = WithContext(r, ctx)

				go func() {
					select {
					case <-doneChan:
						cancelFunc()
					case <-closeChan:
						cancelFunc()
					}
				}()
			}
			h.ServeHTTP(w, r)
		})
}
