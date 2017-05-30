// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whtmpl provides some helpful utilities for constructing and using
// lots of html/templates
package whtmpl // import "gopkg.in/webhelp.v1/whtmpl"

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/webhelp.v1/wherr"
)

// Pair is a useful type that allows for passing more than one current template
// variable into a sub-template.
//
// Expected usage within a template like:
//
//   {{ template "subtemplate" (makepair $val1 $val2) }}
//
// Expected usage within the subtemplate like
//
//   {{ $val1 := .First }}
//   {{ $val2 := .Second }}
//
// "makepair" is registered as a template function inside a Collection
type Pair struct {
	First, Second interface{}
}

// Collection is a useful type that helps when defining a bunch of html
// inside of Go files. Assuming you want to define a template called "landing"
// that references another template called "header". With a template
// collection, you would make three files:
//
//   pkg.go:
//
//      package views
//
//      import "gopkg.in/webhelp.v1/whtmpl"
//
//      var Templates = whtmpl.NewCollection()
//
//   landing.go:
//
//      package views
//
//      var _ = Templates.MustParse(`{{ template "header" . }}
//
//         <h1>Landing!</h1>`)
//
//   header.go:
//
//      package views
//
//      var _ = Templates.MustParse(`<title>My website!</title>`)
//
// Note that MustParse determines the name of the template based on the
// go filename.
//
// A template collection by default has two additional helper functions defined
// within templates:
//
//  * makemap: creates a map out of the even number of arguments given.
//  * makepair: creates a Pair type of its two given arguments.
//  * makeslice: creates a slice of the given arguments.
//  * safeurl: calls template.URL with its first argument and returns the
//             result.
//
type Collection struct {
	group *template.Template
}

// Creates a new Collection.
func NewCollection() *Collection {
	return &Collection{group: template.New("").Funcs(
		template.FuncMap{
			"makepair": func(first, second interface{}) Pair {
				return Pair{First: first, Second: second}
			},
			"makemap":   makemap,
			"makeslice": func(args ...interface{}) []interface{} { return args },
			"safeurl": func(val string) template.URL {
				return template.URL(val)
			},
			"safehtml": func(val string) template.HTML {
				return template.HTML(val)
			},
		})}
}

func makemap(vals ...interface{}) map[interface{}]interface{} {
	if len(vals)%2 != 0 {
		panic("need an even amount of values")
	}
	rv := make(map[interface{}]interface{}, len(vals)/2)
	for i := 0; i < len(vals); i += 2 {
		rv[vals[i]] = vals[i+1]
	}
	return rv
}

// Allows you to add and overwrite template function definitions. Mutates
// called collection and returns self.
func (tc *Collection) Funcs(m template.FuncMap) *Collection {
	tc.group = tc.group.Funcs(m)
	return tc
}

// MustParse parses template source "tmpl" and stores it in the
// Collection using the name of the go file that MustParse is called
// from.
func (tc *Collection) MustParse(tmpl string) *template.Template {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to determine template name")
	}
	name := strings.TrimSuffix(filepath.Base(filename), ".go")
	parsed, err := tc.Parse(name, tmpl)
	if err != nil {
		panic(err)
	}
	return parsed
}

// Parse parses the source "tmpl" and stores it in the template collection
// using name "name".
func (tc *Collection) Parse(name string, tmpl string) (
	*template.Template, error) {
	if tc.group.Lookup(name) != nil {
		return nil, fmt.Errorf("template %#v already registered", name)
	}

	return tc.group.New(name).Parse(tmpl)
}

// Lookup a template by name. Returns nil if not found.
func (tc *Collection) Lookup(name string) *template.Template {
	return tc.group.Lookup(name)
}

// Render writes the template out to the response writer (or any errors that
// come up), with value as the template value.
func (tc *Collection) Render(w http.ResponseWriter, r *http.Request,
	template string, values interface{}) {
	tmpl := tc.Lookup(template)
	if tmpl == nil {
		wherr.Handle(w, r, wherr.InternalServerError.New(
			"no template %#v registered", template))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.Execute(w, values)
	if err != nil {
		wherr.Handle(w, r, err)
		return
	}
}
