// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"context"
	"net/http"
)

// CloseNotify causes a handler to have its request.Context() get canceled the
// second the client goes away, by hooking the http.CloseNotifier logic
// into the context. I assume the standard library doesn't do this
// automatically due to the small amount of overhead it causes. Without this
// addition, the context will still close when the handler completes.
func CloseNotify(h http.Handler) http.Handler {
	return RouteHandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		if cnw, ok := w.(http.CloseNotifier); ok {
			doneChan := make(chan bool)
			defer close(doneChan)

			closeChan := cnw.CloseNotify()
			ctx, cancelFunc := context.WithCancel(r.Context())
			r = r.WithContext(ctx)

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
