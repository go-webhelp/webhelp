// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jtolds/webhelp"
	"golang.org/x/net/context"
)

func ExampleArgMux(t *testing.T) {
	pageName := webhelp.NewArgMux()
	handler := webhelp.DirMux{
		"wiki": pageName.Shift(webhelp.Exact(webhelp.HandlerFunc(
			func(ctx context.Context, w webhelp.ResponseWriter, r *http.Request) error {
				name := pageName.Get(ctx)
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprintf(w, "Welcome to %s", name)
				return nil
			})))}

	webhelp.ListenAndServe(":0", handler)
}
