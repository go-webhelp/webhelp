// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// +build go1.8

package webhelp

import (
	"net/http"

	"golang.org/x/net/context"
)

// CloseNotify causes a handler to have its request.Context() get canceled the
// second the client goes away, by hooking the http.CloseNotifier logic
// into the context. Without this addition, the context will still close when
// the handler completes. This will no longer be necessary with Go1.8.
func CloseNotify(h http.Handler) http.Handler {
	return h
}
