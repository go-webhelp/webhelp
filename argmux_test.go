// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jtolds/webhelp"
)

func ExampleArgMux(t *testing.T) {
	pageName := webhelp.NewStringArgMux()
	handler := webhelp.DirMux{
		"wiki": pageName.Shift(webhelp.Exact(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				name := pageName.Get(r.Context())
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprintf(w, "Welcome to %s", name)
			})))}

	http.ListenAndServe(":0", handler)
}
