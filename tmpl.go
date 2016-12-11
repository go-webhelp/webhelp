// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"

	"strings"
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
// "makepair" is registered as a template function inside TemplateCollections
type Pair struct {
	First, Second interface{}
}

// TemplateCollection is a useful type that helps when defining a bunch of html
// inside of Go files. Assuming you want to define a template called "landing"
// that references another template called "header". With a template
// collection, you would make three files:
//
//   pkg.go:
//
//      package views
//
//      import "github.com/jtolds/webhelp"
//
//      var Templates = webhelp.NewTemplateCollection()
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
//  * makepair: creates a Pair type of its two given arguments
//  * safeurl: calls template.URL with its first argument and returns the
//             result
//
type TemplateCollection struct {
	group *template.Template
}

// Creates a new TemplateCollection.
func NewTemplateCollection() *TemplateCollection {
	return &TemplateCollection{group: template.New("").Funcs(
		template.FuncMap{
			"makepair": func(first, second interface{}) Pair {
				return Pair{First: first, Second: second}
			},
			"safeurl": func(val string) template.URL {
				return template.URL(val)
			},
		})}
}

// Allows you to add and overwrite template function definitions.
func (tc *TemplateCollection) Funcs(m template.FuncMap) {
	tc.group = tc.group.Funcs(m)
}

// MustParse parses template source "tmpl" and stores it in the
// TemplateCollection using the name of the go file that MustParse is called
// from.
func (tc *TemplateCollection) MustParse(tmpl string) *template.Template {
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
func (tc *TemplateCollection) Parse(name string, tmpl string) (
	*template.Template, error) {
	if tc.group.Lookup(name) != nil {
		return nil, fmt.Errorf("template %#v already registered", name)
	}

	return tc.group.New(name).Parse(tmpl)
}

// Lookup a template by name. Returns nil if not found.
func (tc *TemplateCollection) Lookup(name string) *template.Template {
	return tc.group.Lookup(name)
}

// Render writes the template out to the response writer (or any errors that
// come up), with value as the template value.
func (tc *TemplateCollection) Render(w http.ResponseWriter, r *http.Request,
	template string, values interface{}) {
	tmpl := tc.Lookup(template)
	if tmpl == nil {
		HandleError(w, r, ErrInternalServerError.New(
			"no template %#v registered", template))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.Execute(w, values)
	if err != nil {
		HandleError(w, r, err)
		return
	}
}
