// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"fmt"
	"io"
)

const (
	AllMethods = "ALL"
	AllPaths   = "[/<*>]"
)

type RouteLister interface {
	Routes(cb func(method, path string, annotations []string))
}

func Routes(h Handler, cb func(method, path string, annotations []string)) {
	if rl, ok := h.(RouteLister); ok {
		rl.Routes(cb)
	} else {
		cb(AllMethods, AllPaths, nil)
	}
}

func PrintRoutes(out io.Writer, h Handler) (err error) {
	Routes(h, func(method, path string, annotations []string) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(out, "%s %s\n", method, path)
		for _, annotation := range annotations {
			if err != nil {
				return
			}
			_, err = fmt.Fprintf(out, " %s\n", annotation)
		}
	})
	return err
}
