// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
	"golang.org/x/net/context"
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

	errHandler = errors.GenSym()
)

// Redirect is just http.Redirect with http.StatusSeeOther which I always
// forget.
func Redirect(w http.ResponseWriter, r *http.Request, redirectTo string) {
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// HandleError uses the provided error handler given via HandleErrorsWith
// to handle the error, falling back to a built in default if not provided.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	handler, ok := Context(r).Value(errHandler).(ErrorHandler)
	if ok {
		handler.HandleError(w, r, err)
		return
	}
	logger.Errorf("error: %v", err)
	http.Error(w, errhttp.GetErrorBody(err),
		errhttp.GetStatusCode(err, http.StatusInternalServerError))
}

type ErrorHandler interface {
	HandleError(w http.ResponseWriter, r *http.Request, err error)
}

func HandleErrorsWith(eh ErrorHandler, h http.Handler) http.Handler {
	return RouteHandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(Context(r), errHandler, eh)
		h.ServeHTTP(w, WithContext(r, ctx))
	})
}

type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

func (f ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request,
	err error) {
	f(w, r, err)
}

var (
	JSONErrorHandler = ErrorHandlerFunc(jsonErrorHandler)
)

func jsonErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	logger.Errorf("error: %v", err)
	data, err := json.MarshalIndent(map[string]string{
		"err": errhttp.GetErrorBody(err)}, "", "  ")
	if err != nil {
		logger.Critf("failed serializing error: %v", err)
		data = []byte(`{"err": "Internal Server Error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	w.WriteHeader(errhttp.GetStatusCode(err, http.StatusInternalServerError))
	w.Write(data)
}
