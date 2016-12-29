// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package wherr_test

import (
	"fmt"
	"net/http"

	"gopkg.in/webhelp.v1/wherr"
	"gopkg.in/webhelp.v1/whlog"
	"gopkg.in/webhelp.v1/whmux"
	"github.com/spacemonkeygo/errors/errhttp"
)

func PageName(r *http.Request) (string, error) {
	if r.FormValue("name") == "" {
		return "", wherr.BadRequest.New("No page name supplied")
	}
	return r.FormValue("name"), nil
}

func Page(w http.ResponseWriter, r *http.Request) {
	name, err := PageName(r)
	if err != nil {
		// This will use our error handler!
		wherr.Handle(w, r, err)
		return
	}

	fmt.Fprintf(w, name)
	// do more stuff
}

func Routes() http.Handler {
	return whmux.Dir{
		"page": http.HandlerFunc(Page),
	}
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "some error happened!", errhttp.GetStatusCode(err, 500))
}

func Example() {
	// If we didn't register our error handler, we'd end up using a default one.
	whlog.ListenAndServe(":0", wherr.HandleWith(wherr.HandlerFunc(ErrorHandler),
		Routes()))
}
