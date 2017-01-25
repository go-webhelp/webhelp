// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whauth provides some helper methods and handlers for dealing with
// HTTP basic auth
package whauth // import "gopkg.in/webhelp.v1/whauth"

import (
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/webhelp.v1"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whroute"
)

var (
	BasicAuthUser = webhelp.GenSym()
)

// RequireBasicAuth ensures that a valid user is provided, calling
// wherr.Handle with wherr.Unauthorized if not.
func RequireBasicAuth(h http.Handler, realm string,
	valid func(user, pass string) bool) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				wherr.Handle(w, r, wherr.Unauthorized.New("basic auth required"))
				return
			}
			if !valid(user, pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				wherr.Handle(w, r,
					wherr.Unauthorized.New("invalid username or password"))
				return
			}
			h.ServeHTTP(w, whcompat.WithContext(r, context.WithValue(
				whcompat.Context(r), BasicAuthUser, user)))
		})
}
