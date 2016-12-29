// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package wherr provides a unified error handling framework for http.Handlers.
package wherr // import "gopkg.in/webhelp.v1/wherr"

import (
	"log"
	"net/http"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
	"golang.org/x/net/context"
	"gopkg.in/webhelp.v1"
	"gopkg.in/webhelp.v1/whcompat"
	"gopkg.in/webhelp.v1/whroute"
)

var (
	HTTPError = errors.NewClass("HTTP Error", errors.NoCaptureStack())

	BadRequest                   = ErrorClass(http.StatusBadRequest)
	Unauthorized                 = ErrorClass(http.StatusUnauthorized)
	Forbidden                    = ErrorClass(http.StatusForbidden)
	NotFound                     = ErrorClass(http.StatusNotFound)
	MethodNotAllowed             = ErrorClass(http.StatusMethodNotAllowed)
	NotAcceptable                = ErrorClass(http.StatusNotAcceptable)
	RequestTimeout               = ErrorClass(http.StatusRequestTimeout)
	Conflict                     = ErrorClass(http.StatusConflict)
	Gone                         = ErrorClass(http.StatusGone)
	LengthRequired               = ErrorClass(http.StatusLengthRequired)
	PreconditionFailed           = ErrorClass(http.StatusPreconditionFailed)
	RequestEntityTooLarge        = ErrorClass(http.StatusRequestEntityTooLarge)
	RequestURITooLong            = ErrorClass(http.StatusRequestURITooLong)
	UnsupportedMediaType         = ErrorClass(http.StatusUnsupportedMediaType)
	RequestedRangeNotSatisfiable = ErrorClass(http.StatusRequestedRangeNotSatisfiable)
	ExpectationFailed            = ErrorClass(http.StatusExpectationFailed)
	Teapot                       = ErrorClass(http.StatusTeapot)
	InternalServerError          = ErrorClass(http.StatusInternalServerError)
	NotImplemented               = ErrorClass(http.StatusNotImplemented)
	BadGateway                   = ErrorClass(http.StatusBadGateway)
	ServiceUnavailable           = ErrorClass(http.StatusServiceUnavailable)
	GatewayTimeout               = ErrorClass(http.StatusGatewayTimeout)

	errHandler = webhelp.GenSym()
)

// ErrorClass creates a new subclass of HTTPError using the given HTTP status
// code
func ErrorClass(code int) *errors.ErrorClass {
	msg := http.StatusText(code)
	if msg == "" {
		msg = "Unknown error"
	}
	return HTTPError.NewClass(msg, errhttp.SetStatusCode(code))
}

// Handle uses the provided error handler given via HandleWith
// to handle the error, falling back to a built in default if not provided.
func Handle(w http.ResponseWriter, r *http.Request, err error) {
	if handler, ok := whcompat.Context(r).Value(errHandler).(Handler); ok {
		handler.HandleError(w, r, err)
		return
	}
	log.Printf("error: %v", err)
	http.Error(w, errhttp.GetErrorBody(err),
		errhttp.GetStatusCode(err, http.StatusInternalServerError))
}

// Handlers handle errors. After HandleError returns, it's assumed a response
// has been written out and all error handling has completed.
type Handler interface {
	HandleError(w http.ResponseWriter, r *http.Request, err error)
}

// HandleWith binds the given eror Handler to the request contexts that pass
// through the given http.Handler. wherr.Handle will use this error Handler
// for handling errors. If you're using the whfatal package, you should place
// a whfatal.Catch inside this handler, so this error handler can deal
// with Fatal requests.
func HandleWith(eh Handler, h http.Handler) http.Handler {
	return whroute.HandlerFunc(h,
		func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(whcompat.Context(r), errHandler, eh)
			h.ServeHTTP(w, whcompat.WithContext(r, ctx))
		})
}

// HandlingWith returns the error handler if registered, or nil if no error
// handler is registered and the default should be used.
func HandlingWith(ctx context.Context) Handler {
	handler, _ := ctx.Value(errHandler).(Handler)
	return handler
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

func (f HandlerFunc) HandleError(w http.ResponseWriter, r *http.Request,
	err error) {
	f(w, r, err)
}
