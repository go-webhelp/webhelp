// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"context"
	"net/http"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
)

type genSym int

var (
	errHandler   genSym = 1
	stopHandling genSym = 2
)

var (
	HTTPError     = errors.NewClass("HTTP Error", errors.NoCaptureStack())
	ErrBadRequest = HTTPError.NewClass("Bad request",
		errhttp.SetStatusCode(http.StatusBadRequest))
	ErrNotFound = ErrBadRequest.NewClass("Not found",
		errhttp.SetStatusCode(http.StatusNotFound))
	ErrMethodNotAllowed = ErrBadRequest.NewClass("Method not allowed",
		errhttp.SetStatusCode(http.StatusMethodNotAllowed))
	ErrInternalServerError = HTTPError.NewClass("Internal server error",
		errhttp.SetStatusCode(http.StatusInternalServerError))
	ErrUnauthorized = HTTPError.NewClass("Unauthorized",
		errhttp.SetStatusCode(http.StatusUnauthorized))
)

// Redirect is just http.Redirect with http.StatusSeeOther which I always
// forget.
func Redirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// FatalRedirect is like Redirect but panics so all additional request
// processing terminates. LoggingHandler understands this sort of panic and
// squashes it.
func FatalRedirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	Redirect(w, r, redirectTo)
	panic(stopHandling)
}

// HandleError uses the provided error handler given via HandleErrorsWith
// to handle the error, falling back to a built in default if not provided.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	handler, ok := r.Context().Value(errHandler).(ErrorHandler)
	if ok {
		handler.HandleError(w, r, err)
		return
	}
	http.Error(w, errhttp.GetErrorBody(err),
		errhttp.GetStatusCode(err, http.StatusInternalServerError))
}

// FatalError is like HandleError but panics so that all additional request
// processing terminates. LoggingHandler understands this sort of panic and
// squashes it.
func FatalError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err)
	panic(stopHandling)
}

type ErrorHandler interface {
	HandleError(w http.ResponseWriter, r *http.Request, err error)
}

func HandleErrorsWith(eh ErrorHandler, h http.Handler) http.Handler {
	return RouteHandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), errHandler, eh)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
