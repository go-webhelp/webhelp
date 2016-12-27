// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp_test

import (
	"fmt"
	"net/http"

	"github.com/jtolds/webhelp"
	"github.com/jtolds/webhelp/whcompat"
	"github.com/jtolds/webhelp/wherr"
	"github.com/jtolds/webhelp/whlog"
	"github.com/jtolds/webhelp/whroute"
	"golang.org/x/net/context"
)

var (
	UserKey = webhelp.GenSym()
)

type User struct {
	Name string
}

func loadUser(r *http.Request) (user *User, err error) {
	return nil, wherr.InternalServerError.New("not implemented yet")
}

// myWrapper will load the user from a request, serving any detected errors,
// and otherwise passing the request along to the wrapped handler with the
// user bound inside the context.
func myWrapper(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {

			user, err := loadUser(r)
			if err != nil {
				wherr.Handle(w, r, err)
				return
			}

			h.ServeHTTP(w, whcompat.WithContext(r,
				context.WithValue(whcompat.Context(r), UserKey, user)))
		})
}

// myHandler is a standard http.HandlerFunc that expects to be able to load
// a user out of the request context.
func myHandler(w http.ResponseWriter, r *http.Request) {
	ctx := whcompat.Context(r)
	if user, ok := ctx.Value(UserKey).(*User); ok {
		// do something with the user
		fmt.Fprint(w, user.Name)
	}
}

// Routes returns an http.Handler. You might have a whmux.Dir or something
// in here.
func Routes() http.Handler {
	return myWrapper(http.HandlerFunc(myHandler))
}

func ExampleGenSym() {
	whlog.ListenAndServe(":0", Routes())
}
