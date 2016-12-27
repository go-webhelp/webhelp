// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// +build appengine
// +build !go1.7

package whcompat

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

func new16Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}
