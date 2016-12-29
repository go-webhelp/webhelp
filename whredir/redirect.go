// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whredir provides some helper methods and handlers for redirecting
// incoming requests to other URLs.
package whredir // import "gopkg.in/webhelp.v1/whredir"

import (
	"net/http"

	"gopkg.in/webhelp.v1/whmux"
	"gopkg.in/webhelp.v1/whroute"
)

// Redirect is just http.Redirect with http.StatusSeeOther which I always
// forget.
func Redirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

type redirectHandler string

// RedirectHandler returns an http.Handler that redirects all requests to url.
func RedirectHandler(url string) http.Handler {
	return redirectHandler(url)
}

// ServeHTTP implements http.handler
func (t redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Redirect(w, r, string(t))
}

// Routes implements whroute.Lister
func (t redirectHandler) Routes(
	cb func(method, path string, annotations map[string]string)) {
	cb(whroute.AllMethods, whroute.AllPaths,
		map[string]string{"Redirect": string(t)})
}

var _ http.Handler = redirectHandler("")
var _ whroute.Lister = redirectHandler("")

// RedirectHandlerFunc is an http.Handler that redirects all requests to the
// returned URL.
type RedirectHandlerFunc func(r *http.Request) string

// ServeHTTP implements http.handler
func (f RedirectHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Redirect(w, r, f(r))
}

// Routes implements whroute.Lister
func (f RedirectHandlerFunc) Routes(
	cb func(method, path string, annotations map[string]string)) {
	cb(whroute.AllMethods, whroute.AllPaths,
		map[string]string{"Redirect": "f(req)"})
}

var _ http.Handler = RedirectHandlerFunc(nil)
var _ whroute.Lister = RedirectHandlerFunc(nil)

// RequireHTTPS returns a handler that will redirect to the same path but using
// https if https was not already used.
func RequireHTTPS(handler http.Handler) http.Handler {
	return whroute.HandlerFunc(handler,
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Scheme != "https" {
				u := *r.URL
				u.Scheme = "https"
				Redirect(w, r, u.String())
			} else {
				handler.ServeHTTP(w, r)
			}
		})
}

// RequireHost returns a handler that will redirect to the same path but using
// the given host if the given host was not specifically requested.
func RequireHost(host string, handler http.Handler) http.Handler {
	if host == "*" {
		return handler
	}
	return whmux.Host{
		host: handler,
		"*": RedirectHandlerFunc(func(r *http.Request) string {
			u := *r.URL
			u.Host = host
			return u.String()
		})}
}
