// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp_test

import (
	"fmt"
	"net/http"

	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/whlog"
	"gopkg.in/webhelp.v1/whmux"
)

var (
	pageName = whmux.NewStringArg()
)

func page(w http.ResponseWriter, r *http.Request) {
	name := pageName.Get(whcompat.Context(r))

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Welcome to %s", name)
}

func Example() {
	pageHandler := pageName.Shift(whmux.Exact(http.HandlerFunc(page)))

	whlog.ListenAndServe(":0", whmux.Dir{
		"wiki": pageHandler,
	})
}
