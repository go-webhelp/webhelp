// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package webhelp

import (
	"net/http"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
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

func Redirect(w ResponseWriter, r *http.Request, redirectTo string) error {
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	return nil
}
